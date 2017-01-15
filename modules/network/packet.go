package network

import (
	"encoding/binary"
	"fmt"
)

const (
	onlineCheckID                     PacketNameID = 0x1001
	disconnectionID                   PacketNameID = 0x1010
	reconnectionID                    PacketNameID = 0x1011
	fastDataID                        PacketNameID = 0x1020
	connectionSummaryID               PacketNameID = 0x1021
	authenticationInformationHeaderID PacketNameID = 0x1101
	authenticationInformationDataID   PacketNameID = 0x1102
	authenticationInformationFooterID PacketNameID = 0x1103
	tusCommonAreaAcquisitionID        PacketNameID = 0x1201
	tusCommonAreaSettingsID           PacketNameID = 0x1202
	tusCommonAreaAddID                PacketNameID = 0x1203
	tusUserAreaWriteHeaderID          PacketNameID = 0x1204
	tusUserAreaWriteDataID            PacketNameID = 0x1205
	tusUserAreaWriteFooterID          PacketNameID = 0x1206
	tusUserAreaReadHeaderID           PacketNameID = 0x1207
	tusUserAreaReadDataID             PacketNameID = 0x1208
	tusUserAreaReadFooterID           PacketNameID = 0x1209
	unknownNameID                     PacketNameID = 0xFFFF
)

const (
	requestID      PacketTypeID = 0x01
	responseID     PacketTypeID = 0x02
	notificationID PacketTypeID = 0x10
	unknownTypeID  PacketTypeID = 0xFF
)

const (
	noErrorID      PacketErrorID = 0x00
	unknownErrorID PacketErrorID = 0xFF
)

const (
	maxPacketLength = int(^uint16(0))
)

type Packet interface {
	Payload() ([]byte, error)
	Type() PacketType
	SetHeader(header PacketHeader)
	SetPayload(payload []byte) error
}

type PacketNameID uint16

type PacketTypeID uint8

type PacketErrorID uint8

type PacketType struct {
	NameID  PacketNameID
	TypeID  PacketTypeID
	ErrorID PacketErrorID
}

// packet content

type Property struct {
	Index  uint8
	Value1 uint32
	Value2 uint32
}

// base packets

type PacketHeader struct {
	Length     uint16
	SequenceID uint16
	PacketType PacketType
}

func (p *PacketHeader) String() string {
	return p.PacketType.String()
}

func (p *PacketHeader) Payload() ([]byte, error) {
	var payload [0]byte
	return payload[:], nil
}

func (p *PacketHeader) Type() PacketType {
	return PacketType{unknownNameID, unknownTypeID, unknownErrorID}
}

func (p *PacketHeader) SetHeader(header PacketHeader) {
	p.Length = header.Length
	p.SequenceID = header.SequenceID
	p.PacketType = header.PacketType
}

func (p *PacketHeader) SetPayload(payload []byte) error {
	if len(payload) > 0 {
		return PayloadError{len(payload), 0}
	}

	return nil
}

type EmptyPacket struct {
	PacketHeader
}

func (p *EmptyPacket) SetPayload(payload []byte) error {
	const expectedPacketSize = 0
	if len(payload) != 0 {
		return NewPayloadError(len(payload), expectedPacketSize)
	}

	return nil
}

type BooleanPacket struct {
	PacketHeader
	Value bool
}

func (p *BooleanPacket) Payload() ([]byte, error) {
	var payload [1]byte

	if p.Value {
		payload[0] = 1
	}

	return payload[:], nil
}

type DataChunkPacket struct {
	PacketHeader
	ChunkOffset uint32
	ChunkLength uint16
	ChunkData   []byte
}

func (p *DataChunkPacket) SetPayload(payload []byte) error {
	const headerLength = 6
	if len(payload) < headerLength {
		return NewPayloadError(len(payload), headerLength)
	}

	p.ChunkOffset = binary.BigEndian.Uint32(payload[0:4])
	p.ChunkLength = binary.BigEndian.Uint16(payload[4:6])
	p.ChunkData = payload[6:]
	return nil
}

type DataChunkReferencePacket struct {
	PacketHeader
	ChunkOffset uint32
	ChunkLength uint16
}

func (p *DataChunkReferencePacket) Payload() ([]byte, error) {
	var payload [6]byte

	binary.BigEndian.PutUint32(payload[0:4], p.ChunkOffset)
	binary.BigEndian.PutUint16(payload[4:6], p.ChunkLength)

	return payload[:], nil
}

type PropertyPacket struct {
	PacketHeader
	Properties []Property
}

// requests

type OnlineCheckRequest struct {
	EmptyPacket
}

type DisconnectionRequest struct {
	BooleanPacket
}

type FastDataRequest struct {
	EmptyPacket
}

func (p FastDataRequest) Type() PacketType {
	return PacketType{fastDataID, requestID, noErrorID}
}

type AuthenticationInformationRequestHeader struct {
	PacketHeader
	_          uint8 // 0x02
	DataLength uint32
}

func (p *AuthenticationInformationRequestHeader) SetPayload(payload []byte) error {
	const headerLength = 5
	if len(payload) != headerLength {
		return NewPayloadError(len(payload), headerLength)
	}

	p.DataLength = binary.BigEndian.Uint32(payload[:])
	return nil
}

type AuthenticationInformationRequestData struct {
	DataChunkPacket
}

type AuthenticationInformationRequestFooter struct {
	EmptyPacket
}

type TusCommonAreaAcquisitionRequest struct {
	PacketHeader
	PropertyIndices []byte
}

type TusCommonAreaSettingsRequest struct {
	PropertyPacket
}

type TusCommonAreaAddRequest struct {
	PropertyPacket
}

type TusUserAreaWriteRequestHeader struct {
	PacketHeader
	DataLength uint32
	UserLength uint16
	User       string
}

type TusUserAreaWriteRequestData struct {
	DataChunkPacket
}

type TusUserAreaWriteRequestFooter struct {
	EmptyPacket
}

type TusUserAreaReadRequestHeader struct {
	PacketHeader
	UserLength uint16
	User       string
}

type TusUserAreaReadRequestData struct {
	DataChunkReferencePacket
}

type TusUserAreaReadRequestFooter struct {
	EmptyPacket
}

// responses

type OnlineCheckResponse struct {
	EmptyPacket
}

type DisconnectionResponse struct {
	BooleanPacket
}

type FastDataResponse struct {
	PacketHeader
	_    uint8  // 0x03
	_    uint32 // 0x00000001
	_    uint16 // len(User)
	User string
}

func (p *FastDataResponse) SetPayload(payload []byte) error {
	const headerLength = 7
	if len(payload) < headerLength {
		return NewPayloadError(len(payload), headerLength)
	}

	p.User = string(payload[7:])
	return nil
}

type AuthenticationInformationResponseHeader struct {
	PacketHeader
	ChunkLength uint16
}

func (p *AuthenticationInformationResponseHeader) Type() PacketType {
	return PacketType{authenticationInformationHeaderID, responseID, noErrorID}
}

func (p *AuthenticationInformationResponseHeader) Payload() ([]byte, error) {
	var payload [2]byte

	binary.BigEndian.PutUint16(payload[:], p.ChunkLength)

	return payload[:], nil
}

type AuthenticationInformationResponseData struct {
	DataChunkReferencePacket
}

func (p *AuthenticationInformationResponseData) Type() PacketType {
	return PacketType{authenticationInformationDataID, responseID, noErrorID}
}

type AuthenticationInformationResponseFooter struct {
	BooleanPacket
}

func (p *AuthenticationInformationResponseFooter) Type() PacketType {
	return PacketType{authenticationInformationFooterID, responseID, noErrorID}
}

type TusCommonAreaAcquisitionResponse struct {
	PropertyPacket
}

type TusCommonAreaSettingsResponse struct {
	PropertyPacket
}

type TusCommonAreaAddResponse struct {
	PropertyPacket
}

type TusUserAreaWriteResponseHeader struct {
	PacketHeader
	ChunkLength uint16
}

type TusUserAreaWriteResponseData struct {
	DataChunkReferencePacket
}

type TusUserAreaWriteResponseFooter struct {
	EmptyPacket
}

type TusUserAreaReadResponseHeader struct {
	PacketHeader
	DataLength uint32
}

type TusUserAreaReadResponseData struct {
	DataChunkPacket
}

type TusUserAreaReadResponseFooter struct {
	EmptyPacket
}

// notifications

type DisconnectionNotification struct {
	PacketHeader
	Unknown      byte
	_            uint16 // len(Notification)
	Notification string
}

func (p *DisconnectionNotification) Type() PacketType {
	return PacketType{disconnectionID, notificationID, noErrorID}
}

func (p *DisconnectionNotification) Payload() ([]byte, error) {
	notificationPayload := []byte(p.Notification)
	if len(notificationPayload) > maxPacketLength {
		return nil, NewPayloadError(len(notificationPayload), maxPacketLength)
	}

	var headerPayload [3]byte
	headerPayload[0] = p.Unknown
	binary.BigEndian.PutUint16(headerPayload[1:3], uint16(len(notificationPayload)))

	payload := append(headerPayload[:], notificationPayload...)
	return payload[:], nil
}

type ReconnectionNotification struct {
	PacketHeader
	Host string
	Port uint16
}

type ConnectionSummaryNotification struct {
	PacketHeader
	Success bool
	Unknown uint16
}

func (p *ConnectionSummaryNotification) Type() PacketType {
	return PacketType{connectionSummaryID, notificationID, noErrorID}
}

func (p *ConnectionSummaryNotification) Payload() ([]byte, error) {
	var payload [3]byte

	if p.Success {
		payload[0] = 1
	}

	binary.BigEndian.PutUint16(payload[1:3], p.Unknown)

	return payload[:], nil
}

var (
	packetGenerator map[PacketTypeID]map[PacketNameID]func() Packet
)

func init() {
	packetGenerator = map[PacketTypeID]map[PacketNameID]func() Packet{
		requestID: {
			//onlineCheckID:func() Packet {return &OnlineCheckRequest{}},
			//disconnectionID:func() Packet {return &DisconnectionRequest{}},
			//reconnectionID:func() Packet {return &ReconnectionRequest{}},
			//fastDataID: func() Packet { return &FastDataRequest{} },
			//connectionSummaryID:func() Packet {return &ConnectionSummaryRequest{}},
			authenticationInformationHeaderID: func() Packet { return &AuthenticationInformationRequestHeader{} },
			authenticationInformationDataID:   func() Packet { return &AuthenticationInformationRequestData{} },
			authenticationInformationFooterID: func() Packet { return &AuthenticationInformationRequestFooter{} },
			//tusCommonAreaAcquisitionID:func() Packet {return &TusCommonAreaAcquisitionRequest{}},
			//tusCommonAreaSettingsID:func() Packet {return &TusCommonAreaSettingsRequest{}},
			//tusCommonAreaAddID:func() Packet {return &TusCommonAreaAddRequest{}},
			//tusUserAreaWriteHeaderID:func() Packet {return &TusUserAreaWriteHeaderRequest{}},
			//tusUserAreaWriteDataID:func() Packet {return &TusUserAreaWriteDataRequest{}},
			//tusUserAreaWriteFooterID:func() Packet {return &TusUserAreaWriteFooterRequest{}},
			//tusUserAreaReadHeaderID:func() Packet {return &TusUserAreaReadHeaderRequest{}},
			//tusUserAreaReadDataID:func() Packet {return &TusUserAreaReadDataRequest{}},
			//tusUserAreaReadFooterID:func() Packet {return &TusUserAreaReadFooterRequest{}},
		},
		responseID: {
			//onlineCheckID:func() Packet {return &OnlineCheckResponse{}},
			//disconnectionID:func() Packet {return &DisconnectionResponse{}},
			//reconnectionID:func() Packet {return &ReconnectionResponse{}},
			fastDataID: func() Packet { return &FastDataResponse{} },
			//connectionSummaryID:func() Packet {return &ConnectionSummaryResponse{}},
			//authenticationInformationHeaderID:func() Packet {return &AuthenticationInformationHeaderResponse{}},
			//authenticationInformationDataID:func() Packet {return &AuthenticationInformationDataResponse{}},
			//authenticationInformationFooterID:func() Packet {return &AuthenticationInformationFooterResponse{}},
			//tusCommonAreaAcquisitionID:func() Packet {return &TusCommonAreaAcquisitionResponse{}},
			//tusCommonAreaSettingsID:func() Packet {return &TusCommonAreaSettingsResponse{}},
			//tusCommonAreaAddID:func() Packet {return &TusCommonAreaAddResponse{}},
			//tusUserAreaWriteHeaderID:func() Packet {return &TusUserAreaWriteHeaderResponse{}},
			//tusUserAreaWriteDataID:func() Packet {return &TusUserAreaWriteDataResponse{}},
			//tusUserAreaWriteFooterID:func() Packet {return &TusUserAreaWriteFooterResponse{}},
			//tusUserAreaReadHeaderID:func() Packet {return &TusUserAreaReadHeaderResponse{}},
			//tusUserAreaReadDataID:func() Packet {return &TusUserAreaReadDataResponse{}},
			//tusUserAreaReadFooterID:func() Packet {return &TusUserAreaReadFooterResponse{}},
		},
		notificationID: {
		//disconnectionID:func() Packet {return &DisconnectionNotification{}},
		//reconnectionID:func() Packet {return &ReconnectionNotification{}},
		//connectionSummaryID:func() Packet {return &ConnectionSummaryNotification{}},
		},
	}
}

func (pt *PacketType) String() string {
	var n string
	switch pt.NameID {
	case onlineCheckID:
		n = "onlineCheck"
	case disconnectionID:
		n = "disconnection"
	case reconnectionID:
		n = "reconnection"
	case fastDataID:
		n = "fastData"
	case connectionSummaryID:
		n = "connectionSummary"
	case authenticationInformationHeaderID:
		n = "authenticationInformationHeader"
	case authenticationInformationDataID:
		n = "authenticationInformationData"
	case authenticationInformationFooterID:
		n = "authenticationInformationFooter"
	case tusCommonAreaAcquisitionID:
		n = "tusCommonAreaAcquisition"
	case tusCommonAreaSettingsID:
		n = "tusCommonAreaSettings"
	case tusCommonAreaAddID:
		n = "tusCommonAreaAdd"
	case tusUserAreaWriteHeaderID:
		n = "tusUserAreaWriteHeader"
	case tusUserAreaWriteDataID:
		n = "tusUserAreaWriteData"
	case tusUserAreaWriteFooterID:
		n = "tusUserAreaWriteFooter"
	case tusUserAreaReadHeaderID:
		n = "tusUserAreaReadHeader"
	case tusUserAreaReadDataID:
		n = "tusUserAreaReadData"
	case tusUserAreaReadFooterID:
		n = "tusUserAreaReadFooter"
	default:
		n = fmt.Sprintf("unknown(%x)", pt.NameID)
	}

	var t string
	switch pt.TypeID {
	case requestID:
		t = "request"
	case responseID:
		t = "response"
	case notificationID:
		t = "notification"
	default:
		t = fmt.Sprintf("unknown(%x)", pt.TypeID)
	}

	var e string
	switch pt.ErrorID {
	case noErrorID:
		e = ""
	default:
		e = fmt.Sprintf(" error(%x)", pt.ErrorID)
	}

	return fmt.Sprintf("%s %s%s", n, t, e)
}

func NewPacketFromHeader(header PacketHeader) (Packet, error) {
	packetTypeMap, ok := packetGenerator[header.PacketType.TypeID]
	if !ok {
		return nil, fmt.Errorf("unknwon packet type message type: %x", header.PacketType.TypeID)
	}

	packetFunc, ok := packetTypeMap[header.PacketType.NameID]
	if !ok {
		return nil, fmt.Errorf("unknown packet type: %x", header.PacketType.NameID)
	}

	packet := packetFunc()
	packet.SetHeader(header)
	return packet, nil
}

type PayloadError struct {
	ActualSize   int
	ExpectedSize int
}

func (e PayloadError) Error() string {
	return fmt.Sprintf("invalid payload size %d bytes expected %d bytes", e.ActualSize, e.ExpectedSize)
}

func NewPayloadError(actual int, expected int) PayloadError {
	return PayloadError{actual, expected}
}

type PacketTypeError struct {
	Actual   string
	Expected string
}

func (e PacketTypeError) Error() string {
	return fmt.Sprintf("invalid packet type %s expected %s", e.Actual, e.Expected)
}

func NewPacketTypeError() PacketTypeError {
	return PacketTypeError{}
}
