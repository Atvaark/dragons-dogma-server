package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"time"
)

var localSequenceIDRand = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

func newLocalSequenceID() uint16 {
	return uint16(localSequenceIDRand.Uint32())
}

type ClientConn struct {
	io.ReadWriteCloser
	ID               int64
	User             string
	LocalSequenceID  uint16
	RemoteSequenceID uint16
	ToRemoteClient   bool
}

func NewClientConn(rw io.ReadWriteCloser, ID int64, toRemoteClient bool) *ClientConn {
	return &ClientConn{
		ReadWriteCloser: rw,
		ID:              ID,
		LocalSequenceID: newLocalSequenceID(),
		ToRemoteClient:  toRemoteClient,
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

	if conn.ToRemoteClient && header.PacketType.TypeID == responseID || !conn.ToRemoteClient && header.PacketType.TypeID != responseID {
		conn.LocalSequenceID++
		header.SequenceID = conn.LocalSequenceID
	} else {
		header.SequenceID = conn.RemoteSequenceID
	}

	packet.SetHeader(header)

	printf("%v sending %v\n", conn, &header)

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

	printf("%v receiving %v\n", conn, &header)

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

	conn.RemoteSequenceID = header.SequenceID

	return packet, nil
}
