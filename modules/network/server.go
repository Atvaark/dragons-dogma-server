package network

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	config ServerConfig
}

type ServerConfig struct {
	Port     int
	CertFile string
	KeyFile  string

	tlsConfig tls.Config
}

func NewServer(cfg ServerConfig) (*Server, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, err
	}

	cfg.tlsConfig = tls.Config{
		Certificates: []tls.Certificate{
			cert,
		},
		MaxVersion: tls.VersionTLS10,
	}

	return &Server{
		config: cfg,
	}, nil
}

func (s *Server) ListenAndServe() error {
	port := fmt.Sprintf(":%d", s.config.Port)
	listener, err := tls.Listen("tcp", port, &s.config.tlsConfig)
	if err != nil {
		return err
	}

	var nextConnID int64
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		connID := atomic.AddInt64(&nextConnID, 1)
		go handleConnection(conn, connID)
	}
}

func handleConnection(conn net.Conn, connID int64) {
	defer conn.Close()

	log.Printf("[%d] connecting\n", connID)

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		log.Printf("[%d] no TLS connection\n", connID)
		return
	}

	err := tlsConn.Handshake()
	if err != nil {
		log.Printf("[%d] TLS handshake failed:%v\n", connID, err)
		return
	}

	client, err := authenticate(tlsConn, connID)
	if err != nil {
		log.Printf("[%d] auth failed: %v\n", connID, err)
		return
	}

	log.Printf("%v connected\n", client)

	err = disconnect(client)
	if err != nil {
		log.Printf("%v disconnect failed: %v\n", client, err)
		return
	}

	log.Printf("%v disconnected\n", client)
}

func authenticate(conn *tls.Conn, connID int64) (*ClientConn, error) {
	client := NewClientConn(conn, connID)
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
		return nil, NewPacketTypeError()
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
		return nil, NewPacketTypeError()
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
		return nil, NewPacketTypeError()
	}
	err = client.Send(&AuthenticationInformationResponseData{DataChunkReferencePacket{ChunkLength: authenticationInformationRequestData.ChunkLength}})
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
		return nil, NewPacketTypeError()
	}
	err = client.Send(&AuthenticationInformationResponseFooter{BooleanPacket{Value: true}})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func disconnect(conn *ClientConn) error {
	err := conn.Send(&DisconnectionNotification{})
	if err != nil {
		return err
	}

	return nil
}
