package network

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"

	"io"
)

type ClientConn struct {
	*tls.Conn
	ID               int64
	User             string
	ServerSequenceID uint16
	ClientSequenceID uint16
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
	header.PacketType = packet.Type()
	packet.SetHeader(header)

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

	return packet, nil
}
