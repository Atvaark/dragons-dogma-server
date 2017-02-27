package network

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/atvaark/dragons-dogma-server/modules/game"
)

type Server struct {
	config   ServerConfig
	database game.Database
	listener *serverListener
}

type ServerConfig struct {
	Port      int
	CertFile  string
	KeyFile   string
	tlsConfig *tls.Config
}

func NewServer(cfg ServerConfig, database game.Database) (*Server, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, err
	}
	cfg.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{
			cert,
		},
		MaxVersion: tls.VersionTLS10,
	}

	return &Server{
		config:   cfg,
		database: database,
	}, nil
}

func (s *Server) ListenAndServe() error {
	port := fmt.Sprintf(":%d", s.config.Port)
	tlsListener, err := tls.Listen("tcp", port, s.config.tlsConfig)
	if err != nil {
		return err
	}

	listener := serverListener{
		listener:    tlsListener,
		connections: make(map[int64]net.Conn, 0),
		close:       make(chan bool, 1),
	}
	s.listener = &listener

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-listener.close:
				s.listener.CloseConns()
				return nil
			default:
				printf("%v", err)
				continue
			}
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) Close() error {
	l := s.listener
	if l != nil {
		err := l.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	connID := s.listener.AddConn(conn)
	defer s.listener.DelConn(connID)
	defer conn.Close()

	printf("[%d] connecting\n", connID)

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		printf("[%d] no TLS connection\n", connID)
		return
	}

	err := tlsConn.Handshake()
	if err != nil {
		printf("[%d] TLS handshake failed:%v\n", connID, err)
		return
	}

	client, err := authenticate(tlsConn, connID)
	if err != nil {
		printf("[%d] auth failed: %v\n", connID, err)
		return
	}

	printf("%v connected\n", client)

	err = s.handleClient(client)
	if err != nil {
		printf("%v failed to handle request: %v\n", client, err)
	}

	printf("%v disconnected\n", client)
}

func authenticate(conn *tls.Conn, connID int64) (*ClientConn, error) {
	client := NewClientConn(conn, connID, true)
	var err error
	var response Packet

	err = client.Send(&FastDataRequest{})
	if err != nil {
		return nil, err
	}
	response, err = client.Recv()
	if err != nil {
		return nil, err
	}
	fastDataResponse, ok := response.(*FastDataResponse)
	if !ok {
		return nil, NewPacketTypeError(fastDataResponse, response)
	}

	client.User = fastDataResponse.User

	err = client.Send(&ConnectionSummaryNotification{Success: true, Unknown: 10})
	if err != nil {
		return nil, err
	}

	response, err = client.Recv()
	if err != nil {
		return nil, err
	}
	authenticationInformationRequestHeader, ok := response.(*AuthenticationInformationRequestHeader)
	if !ok {
		return nil, NewPacketTypeError(authenticationInformationRequestHeader, response)
	}
	_ = authenticationInformationRequestHeader
	err = client.Send(&AuthenticationInformationResponseHeader{ChunkLength: 256})
	if err != nil {
		return nil, err
	}

	response, err = client.Recv()
	if err != nil {
		return nil, err
	}
	authenticationInformationRequestData, ok := response.(*AuthenticationInformationRequestData)
	if !ok {
		return nil, NewPacketTypeError(authenticationInformationRequestData, response)
	}
	err = client.Send(&AuthenticationInformationResponseData{DataChunkReferencePacket{ChunkLength: uint16(len(authenticationInformationRequestData.ChunkData))}})
	if err != nil {
		return nil, err
	}

	response, err = client.Recv()
	if err != nil {
		return nil, err
	}
	authenticationInformationRequestFooter, ok := response.(*AuthenticationInformationRequestFooter)
	_ = authenticationInformationRequestFooter
	if !ok {
		return nil, NewPacketTypeError(authenticationInformationRequestFooter, response)
	}
	err = client.Send(&AuthenticationInformationResponseFooter{BooleanPacket{Value: true}})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (s *Server) handleClient(client *ClientConn) error {
	// TODO: Return an error packet to the client in case of errors

	for {
		// TODO: Send an onlineCheckRequest if no request was sent after 60s
		//       Disconnect the client if no onlineCheckResponse gets received after 30s

		request, err := client.Recv()
		if err != nil {
			return err
		}

		switch request := request.(type) {
		case *TusCommonAreaAcquisitionRequest:
			dragon, err := s.database.GetOnlineUrDragon()
			if err != nil {
				return err
			}

			dragonProps, err := dragon.PropertiesFiltered(request.PropertyIndices)
			if err != nil {
				return err
			}

			err = client.Send(&TusCommonAreaAcquisitionResponse{PropertyPacket{Properties: dragonToNetworkProperties(dragonProps)}})
			if err != nil {
				return err
			}
		case *TusCommonAreaAddRequest:
			dragon, err := s.database.GetOnlineUrDragon()
			if err != nil {
				return err
			}

			dragonProps, err := dragon.AddProperties(networkToDragonProperties(request.Properties))
			if err != nil {
				return err
			}

			err = s.database.PutOnlineUrDragon(dragon)
			if err != nil {
				return err
			}

			err = client.Send(&TusCommonAreaAddResponse{PropertyPacket{Properties: dragonToNetworkProperties(dragonProps)}})
			if err != nil {
				return err
			}
		case *TusCommonAreaSettingsRequest:
			dragon, err := s.database.GetOnlineUrDragon()
			if err != nil {
				return err
			}

			err = dragon.SetProperties(networkToDragonProperties(request.Properties))
			if err != nil {
				return err
			}

			err = s.database.PutOnlineUrDragon(dragon)
			if err != nil {
				return err
			}

			err = client.Send(&TusCommonAreaSettingsResponse{PropertyPacket{Properties: request.Properties}})
			if err != nil {
				return err
			}
		case *DisconnectionRequest:
			err = client.Send(&DisconnectionResponse{BooleanPacket{Value: true}})
			if err != nil {
				return err
			}

			return nil
		case *TusUserAreaReadRequestHeader:
			userID, err := strconv.ParseUint(request.User, 16, 64)
			if err != nil {
				return err
			}

			rewards, err := s.database.GetPawnRewards(userID)
			if err != nil {
				return err
			}

			area := pawnRewardsToUserArea(rewards)
			areaData, err := WriteUserArea(area)
			if err != nil {
				return err
			}

			err = client.Send(&TusUserAreaReadResponseHeader{DataLength: uint32(len(areaData))})
			if err != nil {
				return err
			}

		ReadLoop:
			for {
				response, err := client.Recv()
				if err != nil {
					return err
				}

				switch response := response.(type) {
				case *TusUserAreaReadRequestFooter:
					err = client.Send(&TusUserAreaReadResponseFooter{})
					if err != nil {
						return err
					}

					break ReadLoop
				case *TusUserAreaReadRequestData:
					chunkOffset := int(response.ChunkOffset)
					chunkLength := int(response.ChunkLength)

					if chunkLength > maxChunkLength || chunkOffset+chunkLength > len(areaData) {
						return errors.New("read user area failed: invalid size")
					}

					chunkData := areaData[chunkOffset : chunkOffset+chunkLength]
					err = client.Send(&TusUserAreaReadResponseData{DataChunkPacket{ChunkOffset: uint32(chunkOffset), ChunkData: chunkData}})
					if err != nil {
						return err
					}
				default:
					return fmt.Errorf("read user area failed: unexpected response %T", response)
				}
			}
		case *TusUserAreaWriteRequestHeader:
			userID, err := strconv.ParseUint(request.User, 16, 64)
			if err != nil {
				return err
			}

			if request.DataLength > maxDataLength {
				return errors.New("write user area failed: invalid size")
			}

			areaData := make([]byte, int(request.DataLength))

			const chunkLength = uint16(1024)
			err = client.Send(&TusUserAreaWriteResponseHeader{ChunkLength: chunkLength})
			if err != nil {
				return err
			}

		WriteLoop:
			for {
				response, err := client.Recv()
				if err != nil {
					return err
				}

				switch response := response.(type) {
				case *TusUserAreaWriteRequestFooter:
					area, err := ReadUserArea(areaData)
					if err != nil {
						return err
					}

					rewards := userAreaToPawnRewards(userID, area)
					err = s.database.PutPawnRewards(rewards)
					if err != nil {
						return err
					}

					err = client.Send(&TusUserAreaWriteResponseFooter{})
					if err != nil {
						return err
					}

					break WriteLoop
				case *TusUserAreaWriteRequestData:
					chunkOffset := int(response.ChunkOffset)
					chunkData := response.ChunkData
					chunkLength := len(chunkData)

					if chunkLength > maxChunkLength || chunkOffset+chunkLength > len(areaData) {
						return errors.New("write user area failed: invalid size")
					}

					copy(areaData[chunkOffset:chunkOffset+chunkLength], chunkData)

					err = client.Send(&TusUserAreaWriteResponseData{DataChunkReferencePacket{ChunkOffset: uint32(chunkOffset), ChunkLength: uint16(chunkLength)}})
					if err != nil {
						return err
					}
				default:
					return fmt.Errorf("write user area failed: unexpected response %T", response)
				}
			}
		default:
			printf("unhandled request: %v", request)

			err = disconnect(client)
			if err != nil {
				printf("%v disconnect failed: %v\n", client, err)
				return err
			}
		}
	}
}

func disconnect(client *ClientConn) error {
	err := client.Send(&DisconnectionNotification{})
	if err != nil {
		return err
	}

	return nil
}

type serverListener struct {
	listener         net.Listener
	connectionsMutex sync.Mutex
	connections      map[int64]net.Conn
	connId           int64
	close            chan bool
}

func (l *serverListener) Accept() (net.Conn, error) {
	return l.listener.Accept()
}

func (l *serverListener) Close() error {
	l.close <- true
	err := l.listener.Close()
	if err != nil {
		return err
	}

	return nil
}

func (l *serverListener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *serverListener) AddConn(conn net.Conn) (connID int64) {
	l.connectionsMutex.Lock()
	defer l.connectionsMutex.Unlock()
	connID = atomic.AddInt64(&l.connId, 1)
	l.connections[connID] = conn
	return connID
}

func (l *serverListener) DelConn(connID int64) {
	l.connectionsMutex.Lock()
	defer l.connectionsMutex.Unlock()
	delete(l.connections, connID)
}

func (l *serverListener) CloseConns() {
	l.connectionsMutex.Lock()
	defer l.connectionsMutex.Unlock()

	for connID, conn := range l.connections {
		// TODO: Send a DisconnectionNotification Packet to the client before closing the connection.
		err := conn.Close()
		if err != nil {
			printf("[%d] failed to forcefully close connection: %v", connID, err)
		}

		delete(l.connections, connID)
	}
}
