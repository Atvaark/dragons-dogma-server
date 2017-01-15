package network

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"

	"fmt"
	"io"
	"log"
	"math/rand"
	"time"
)

var serverSequenceIDRand = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

func newServerSequenceID() uint16 {
	return uint16(serverSequenceIDRand.Uint32())
}

type ClientConn struct {
	*tls.Conn
	ID               int64
	User             string
	ServerSequenceID uint16
	ClientSequenceID uint16
}

func NewClientConn(conn *tls.Conn, ID int64) *ClientConn {
	return &ClientConn{
		Conn:             conn,
		ID:               ID,
		ServerSequenceID: newServerSequenceID(),
	}
}

func (conn *ClientConn) String() string {
	if len(conn.User) > 0 {
		return fmt.Sprintf("[%d/%s]", conn.ID, conn.User)
	}

	return fmt.Sprintf("[%d]", conn.ID)
}

func (conn *ClientConn) Send(packet Packet) error {
	payload, err := packet.Payload()
	if err != nil {
		return err
	}

	packetLength := len(payload)
	if packetLength > maxPacketLength {
		return NewPayloadError(packetLength, maxPacketLength)
	}

	var header PacketHeader
	header.Length = uint16(packetLength)
	header.PacketType = GetPacketType(packet)

	switch header.PacketType.TypeID {
	case responseID:
		header.SequenceID = conn.ClientSequenceID
	case requestID:
	case notificationID:
		conn.ServerSequenceID++
		header.SequenceID = conn.ServerSequenceID
	}

	packet.SetHeader(header)

	log.Printf("%v sending %v\n", conn, &header)

	err = binary.Write(conn, binary.BigEndian, header)
	if err != nil {
		return err
	}

	err = binary.Write(conn, binary.BigEndian, payload)
	if err != nil {
		return err
	}

	return nil
}

func (conn *ClientConn) Recv() (Packet, error) {
	var header PacketHeader
	err := binary.Read(conn, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}

	log.Printf("%v receiving %v\n", conn, &header)

	packet, err := NewPacketFromHeader(header)
	if err != nil {
		return nil, err
	}

	var payloadBuffer bytes.Buffer
	payloadLength := int64(header.Length)
	n, err := io.CopyN(&payloadBuffer, conn, payloadLength)
	if err != nil {
		return nil, err
	}

	if n != payloadLength {
		return nil, NewPayloadError(int(n), int(payloadLength))
	}

	err = packet.SetPayload(payloadBuffer.Bytes())
	if err != nil {
		return nil, err
	}

	conn.ClientSequenceID = header.SequenceID

	return packet, nil
}
