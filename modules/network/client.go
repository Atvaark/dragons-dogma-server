package network

import (
	"crypto/tls"
	"fmt"
	"log"
)

type Client struct {
	cfg ClientConfig
}

type ClientConfig struct {
	Host      string
	Port      int
	User      string
	UserToken []byte
}

func NewClient(cfg ClientConfig) *Client {
	return &Client{
		cfg: cfg,
	}
}

func (c *Client) Connect() error {
	conf := tls.Config{}

	log.Print("connecting\n")
	tlsConn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port), &conf)
	if err != nil {
		return err
	}

	defer tlsConn.Close()

	err = tlsConn.Handshake()
	if err != nil {
		return err
	}

	conn, err := c.authenticate(tlsConn)
	if err != nil {
		return err
	}

	log.Print("connected\n")

	err = handleServer(conn)

	if err != nil {
		return err
	}

	log.Print("disconnected\n")
	return nil
}

func handleServer(conn *ClientConn) error {
	var err error
	var response Packet

	err = conn.Send(&DisconnectionRequest{BooleanPacket{Value: true}})
	if err != nil {
		return err
	}

	response, err = conn.Recv()
	if err != nil {
		return err
	}
	disconnectionResponse, ok := response.(*DisconnectionResponse)
	if !ok {
		return NewPacketTypeError(disconnectionResponse, response)
	}

	return nil
}

func (c *Client) authenticate(tlsConn *tls.Conn) (*ClientConn, error) {
	conn := NewClientConn(tlsConn, 0, false)
	var err error
	var response Packet

	response, err = conn.Recv()
	if err != nil {
		return nil, err
	}
	fastDataRequest, ok := response.(*FastDataRequest)
	if !ok {
		return nil, NewPacketTypeError(fastDataRequest, response)
	}
	_ = fastDataRequest
	err = conn.Send(&FastDataResponse{Unknown1: 0x03, Unknown2: 0x01, User: c.cfg.User})
	if err != nil {
		return nil, err
	}

	response, err = conn.Recv()
	if err != nil {
		return nil, err
	}
	connectionSummaryNotification, ok := response.(*ConnectionSummaryNotification)
	if !ok {
		return nil, NewPacketTypeError(connectionSummaryNotification, response)
	}
	_ = connectionSummaryNotification

	err = conn.Send(&AuthenticationInformationRequestHeader{Unknown: 0x02, DataLength: uint32(len(c.cfg.UserToken))})
	if err != nil {
		return nil, err
	}
	response, err = conn.Recv()
	if err != nil {
		return nil, err
	}
	authenticationInformationResponseHeader, ok := response.(*AuthenticationInformationResponseHeader)
	if !ok {
		return nil, NewPacketTypeError(authenticationInformationResponseHeader, response)
	}
	_ = authenticationInformationResponseHeader

	err = conn.Send(&AuthenticationInformationRequestData{DataChunkPacket{ChunkData: c.cfg.UserToken}})
	if err != nil {
		return nil, err
	}
	response, err = conn.Recv()
	if err != nil {
		return nil, err
	}
	authenticationInformationResponseData, ok := response.(*AuthenticationInformationResponseData)
	if !ok {
		return nil, NewPacketTypeError(authenticationInformationResponseData, response)
	}
	_ = authenticationInformationResponseData

	err = conn.Send(&AuthenticationInformationRequestFooter{})
	if err != nil {
		return nil, err
	}
	response, err = conn.Recv()
	if err != nil {
		return nil, err
	}
	authenticationInformationResponseFooter, ok := response.(*AuthenticationInformationResponseFooter)
	if !ok {
		return nil, NewPacketTypeError(authenticationInformationResponseFooter, response)
	}
	_ = authenticationInformationResponseFooter

	return conn, nil
}
