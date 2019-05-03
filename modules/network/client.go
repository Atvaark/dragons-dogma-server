package network

import (
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/atvaark/dragons-dogma-server/modules/game"
)

type Client struct {
	cfg  ClientConfig
	conn *ClientConn
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
	if c.conn != nil {
		return nil
	}

	conf := tls.Config{
		InsecureSkipVerify: true, // thawtePrimaryRootCA
	}

	printf("connecting\n")
	tlsConn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port), &conf)
	if err != nil {
		return err
	}

	err = tlsConn.Handshake()
	if err != nil {
		return err
	}

	c.conn = NewClientConn(tlsConn, 0, false)

	printf("authenticating\n")

	err = c.authenticate()
	if err != nil {
		return err
	}

	printf("connected\n")

	return nil
}

func (c *Client) Disconnect() error {
	if c.conn == nil {
		return nil
	}

	var err error
	var response Packet

	err = c.send(&DisconnectionRequest{BooleanPacket{Value: true}})
	if err != nil {
		return err
	}

	response, err = c.recv()
	if err != nil {
		return err
	}
	disconnectionResponse, ok := response.(*DisconnectionResponse)
	if !ok {
		return NewPacketTypeError(disconnectionResponse, response)
	}

	err = c.conn.Close()
	if err != nil {
		return err
	}
	c.conn = nil

	printf("disconnected\n")

	return nil
}

func (c *Client) GetOnlineUrDragon() (*game.OnlineUrDragon, error) {
	if c.conn == nil {
		return nil, errors.New("could not get the online ur dragon. not connected.")
	}

	var err error
	var response Packet

	err = c.send(&TusCommonAreaAcquisitionRequest{PropertyIndices: game.AllDragonPropertyIndices()})
	if err != nil {
		return nil, err
	}

	response, err = c.recv()
	if err != nil {
		return nil, err
	}
	tusCommonAreaAcquisitionResponse, ok := response.(*TusCommonAreaAcquisitionResponse)
	if !ok {
		return nil, NewPacketTypeError(tusCommonAreaAcquisitionResponse, response)
	}

	dragon := &game.OnlineUrDragon{}
	dragon.SetProperties(networkToDragonProperties(tusCommonAreaAcquisitionResponse.Properties))
	return dragon, nil
}

func (c *Client) recv() (Packet, error) {
	if c.conn == nil {
		return nil, errors.New("could not receive data. not connected.")
	}

	return c.conn.Recv()
}

func (c *Client) send(packet Packet) error {
	if c.conn == nil {
		return errors.New("could not send data. not connected.")
	}

	return c.conn.Send(packet)
}

func (c *Client) authenticate() error {
	if c.conn == nil {
		return errors.New("could not authenticate. not connected.")
	}

	var err error
	var response Packet

	response, err = c.recv()
	if err != nil {
		return err
	}
	fastDataRequest, ok := response.(*FastDataRequest)
	if !ok {
		return NewPacketTypeError(fastDataRequest, response)
	}
	err = c.send(&FastDataResponse{Unknown1: 0x03, Unknown2: 0x01, User: c.cfg.User})
	if err != nil {
		return err
	}

	response, err = c.recv()
	if err != nil {
		return err
	}
	connectionSummaryNotification, ok := response.(*ConnectionSummaryNotification)
	if !ok {
		return NewPacketTypeError(connectionSummaryNotification, response)
	}

	err = c.send(&AuthenticationInformationRequestHeader{Unknown: 0x02, DataLength: uint32(len(c.cfg.UserToken))})
	if err != nil {
		return err
	}
	response, err = c.recv()
	if err != nil {
		return err
	}
	authenticationInformationResponseHeader, ok := response.(*AuthenticationInformationResponseHeader)
	if !ok {
		return NewPacketTypeError(authenticationInformationResponseHeader, response)
	}

	err = c.send(&AuthenticationInformationRequestData{DataChunkPacket{ChunkData: c.cfg.UserToken}})
	if err != nil {
		return err
	}
	response, err = c.recv()
	if err != nil {
		return err
	}
	authenticationInformationResponseData, ok := response.(*AuthenticationInformationResponseData)
	if !ok {
		return NewPacketTypeError(authenticationInformationResponseData, response)
	}

	err = c.send(&AuthenticationInformationRequestFooter{})
	if err != nil {
		return err
	}
	response, err = c.recv()
	if err != nil {
		return err
	}
	authenticationInformationResponseFooter, ok := response.(*AuthenticationInformationResponseFooter)
	if !ok {
		return NewPacketTypeError(authenticationInformationResponseFooter, response)
	}

	return nil
}
