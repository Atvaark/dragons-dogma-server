package network

import (
	"encoding/binary"
	"fmt"
	"reflect"
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
	propertySize    = 9
)

type Packet interface {
	SetHeader(header PacketHeader)
	Payload() ([]byte, error)
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

func (p *PacketHeader) SetHeader(header PacketHeader) {
	p.Length = header.Length
	p.SequenceID = header.SequenceID
	p.PacketType = header.PacketType
}

type EmptyPacket struct {
	PacketHeader
}

func (p *EmptyPacket) Payload() ([]byte, error) {
	var payload [0]byte
	return payload[:], nil
}

func (p *EmptyPacket) SetPayload(payload []byte) error {
	const requiredPayloadSize = 0
	if len(payload) != requiredPayloadSize {
		return NewPayloadError(len(payload), requiredPayloadSize)
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

func (p *BooleanPacket) SetPayload(payload []byte) error {
	const requiredPayloadSize = 1
	if len(payload) != requiredPayloadSize {
		return NewPayloadError(len(payload), requiredPayloadSize)
	}

	if payload[0] > 0 {
		p.Value = true
	}

	return nil
}

type DataChunkPacket struct {
	PacketHeader
	ChunkOffset uint32
	_           uint16 //ChunkLength uint16
	_           uint16 //ChunkLength uint16
	ChunkData   []byte
}

func (p *DataChunkPacket) Payload() ([]byte, error) {
	var payload [6]byte
	binary.BigEndian.PutUint32(payload[:4], p.ChunkOffset)
	binary.BigEndian.PutUint16(payload[4:6], uint16(len(p.ChunkData)))

	chunkDataPayload, err := writeDynamicData(p.ChunkData[:])
	if err != nil {
		return nil, err
	}

	return append(payload[:], chunkDataPayload[:]...), nil
}

func (p *DataChunkPacket) SetPayload(payload []byte) error {
	const minPayloadSize = 8
	if len(payload) < minPayloadSize {
		return NewPayloadError(len(payload), minPayloadSize)
	}

	p.ChunkOffset = binary.BigEndian.Uint32(payload[0:4])
	chunkLength := binary.BigEndian.Uint16(payload[4:6])
	_ = chunkLength
	var err error
	p.ChunkData, _, err = readDynamicData(payload[6:])
	if err != nil {
		return err
	}

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

func (p *DataChunkReferencePacket) SetPayload(payload []byte) error {
	const requiredPayloadSize = 6
	if len(payload) != requiredPayloadSize {
		return NewPayloadError(len(payload), requiredPayloadSize)
	}

	p.ChunkOffset = binary.BigEndian.Uint32(payload[:4])
	p.ChunkLength = binary.BigEndian.Uint16(payload[4:6])

	return nil
}

type PropertyPacket struct {
	PacketHeader
	Properties []Property
}

func (p *PropertyPacket) Payload() ([]byte, error) {
	const maxProperties = int(^byte(0))
	if len(p.Properties) > maxProperties {
		return nil, NewPayloadError(len(p.Properties), maxProperties)
	}

	payload := make([]byte, 1+(len(p.Properties)*propertySize))
	payload[0] = byte(len(p.Properties))

	offset := 1
	for _, prop := range p.Properties {
		payload[offset] = prop.Index
		binary.BigEndian.PutUint32(payload[offset+1:offset+5], prop.Value1)
		binary.BigEndian.PutUint32(payload[offset+5:offset+9], prop.Value2)
		offset += propertySize
	}

	return payload, nil
}

func (p *PropertyPacket) SetPayload(payload []byte) error {
	const minPayloadSize = 1
	if len(payload) < minPayloadSize {
		return NewPayloadError(len(payload), minPayloadSize)
	}

	propertyCount := int(payload[0])

	requiredPayloadSize := minPayloadSize + (propertyCount * propertySize)
	if len(payload) < requiredPayloadSize {
		return NewPayloadError(len(payload), requiredPayloadSize)
	}

	p.Properties = make([]Property, propertyCount)
	offset := 1
	for i := 0; i < propertyCount; i++ {
		prop := &p.Properties[i]
		prop.Index = payload[offset]
		prop.Value1 = binary.BigEndian.Uint32(payload[offset+1 : offset+5])
		prop.Value2 = binary.BigEndian.Uint32(payload[offset+5 : offset+9])
		offset += propertySize
	}

	return nil
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

type AuthenticationInformationRequestHeader struct {
	PacketHeader
	Unknown    uint8 // 0x02
	DataLength uint32
}

func (p *AuthenticationInformationRequestHeader) Payload() ([]byte, error) {
	var payload [5]byte
	payload[0] = p.Unknown
	binary.BigEndian.PutUint32(payload[1:], p.DataLength)
	return payload[:], nil
}

func (p *AuthenticationInformationRequestHeader) SetPayload(payload []byte) error {
	const requiredPayloadSize = 5
	if len(payload) != requiredPayloadSize {
		return NewPayloadError(len(payload), requiredPayloadSize)
	}

	p.Unknown = payload[0]
	p.DataLength = binary.BigEndian.Uint32(payload[1:])

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

func (p *TusCommonAreaAcquisitionRequest) Payload() ([]byte, error) {
	const maxIndices = int(^byte(0))
	if len(p.PropertyIndices) > maxIndices {
		return nil, NewPayloadError(len(p.PropertyIndices), maxIndices)
	}

	payload := make([]byte, 1+len(p.PropertyIndices))
	payload[0] = byte(len(p.PropertyIndices))
	copy(payload[1:], p.PropertyIndices[:])

	return payload, nil
}

func (p *TusCommonAreaAcquisitionRequest) SetPayload(payload []byte) error {
	const minPayloadSize = 1
	if len(payload) < minPayloadSize {
		return NewPayloadError(len(payload), minPayloadSize)
	}

	indexCount := int(payload[0])

	requiredPayloadSize := minPayloadSize + indexCount
	if len(payload) < requiredPayloadSize {
		return NewPayloadError(len(payload), requiredPayloadSize)
	}

	p.PropertyIndices = make([]byte, indexCount)
	copy(p.PropertyIndices, payload[1:])

	return nil
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
	_          uint16 // len(User)
	User       string
}

func (p *TusUserAreaWriteRequestHeader) Payload() ([]byte, error) {
	var payload [4]byte
	binary.BigEndian.PutUint32(payload[:], p.DataLength)

	userPayload, err := writeDynamicString(p.User)
	if err != nil {
		return nil, err
	}

	return append(payload[:], userPayload[:]...), nil
}

func (p *TusUserAreaWriteRequestHeader) SetPayload(payload []byte) error {
	const minPayloadSize = 6
	if len(payload) < minPayloadSize {
		return NewPayloadError(len(payload), minPayloadSize)
	}

	p.DataLength = binary.BigEndian.Uint32(payload[:4])

	var err error
	p.User, _, err = readDynamicString(payload[4:])
	if err != nil {
		return err
	}

	return nil
}

type TusUserAreaWriteRequestData struct {
	DataChunkPacket
}

type TusUserAreaWriteRequestFooter struct {
	EmptyPacket
}

type TusUserAreaReadRequestHeader struct {
	PacketHeader
	_    uint16 // len(User)
	User string
}

func (p *TusUserAreaReadRequestHeader) Payload() ([]byte, error) {
	userPayload, err := writeDynamicString(p.User)
	if err != nil {
		return nil, err
	}

	return userPayload[:], nil
}

func (p *TusUserAreaReadRequestHeader) SetPayload(payload []byte) error {
	var err error
	p.User, _, err = readDynamicString(payload[:])
	if err != nil {
		return err
	}

	return nil
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
	Unknown1 uint8  // 0x03
	Unknown2 uint32 // 0x00000001
	_        uint16 // len(User)
	User     string
}

func (p *FastDataResponse) Payload() ([]byte, error) {
	var payload [5]byte
	payload[0] = p.Unknown1
	binary.BigEndian.PutUint32(payload[1:5], p.Unknown2)

	userPayload, err := writeDynamicString(p.User)
	if err != nil {
		return nil, err
	}

	return append(payload[:], userPayload[:]...), nil
}

func (p *FastDataResponse) SetPayload(payload []byte) error {
	const minPayloadSize = 7
	if len(payload) < minPayloadSize {
		return NewPayloadError(len(payload), minPayloadSize)
	}

	var err error
	p.User, _, err = readDynamicString(payload[5:])
	if err != nil {
		return err
	}

	return nil
}

type AuthenticationInformationResponseHeader struct {
	PacketHeader
	ChunkLength uint16
}

func (p *AuthenticationInformationResponseHeader) Payload() ([]byte, error) {
	var payload [2]byte

	binary.BigEndian.PutUint16(payload[:], p.ChunkLength)

	return payload[:], nil
}

func (p *AuthenticationInformationResponseHeader) SetPayload(payload []byte) error {
	const requiredPayloadSize = 2
	if len(payload) != requiredPayloadSize {
		return NewPayloadError(len(payload), requiredPayloadSize)
	}

	p.ChunkLength = binary.BigEndian.Uint16(payload[:])

	return nil
}

type AuthenticationInformationResponseData struct {
	DataChunkReferencePacket
}

type AuthenticationInformationResponseFooter struct {
	BooleanPacket
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

func (p *TusUserAreaWriteResponseHeader) Payload() ([]byte, error) {
	var payload [2]byte
	binary.BigEndian.PutUint16(payload[:], p.ChunkLength)
	return payload[:], nil
}

func (p *TusUserAreaWriteResponseHeader) SetPayload(payload []byte) error {
	const requiredPayloadSize = 2
	if len(payload) != requiredPayloadSize {
		return NewPayloadError(len(payload), requiredPayloadSize)
	}

	p.ChunkLength = binary.BigEndian.Uint16(payload[:])

	return nil
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

func (p *TusUserAreaReadResponseHeader) Payload() ([]byte, error) {
	var payload [4]byte
	binary.BigEndian.PutUint32(payload[:], p.DataLength)
	return payload[:], nil
}

func (p *TusUserAreaReadResponseHeader) SetPayload(payload []byte) error {
	const requiredPayloadSize = 4
	if len(payload) != requiredPayloadSize {
		return NewPayloadError(len(payload), requiredPayloadSize)
	}

	p.DataLength = binary.BigEndian.Uint32(payload[:])

	return nil
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

func (p *DisconnectionNotification) Payload() ([]byte, error) {
	notificationPayload, err := writeDynamicString(p.Notification)
	if err != nil {
		return nil, err
	}

	var headerPayload [1]byte
	headerPayload[0] = p.Unknown
	payload := append(headerPayload[:], notificationPayload[:]...)
	return payload[:], nil
}

func (p *DisconnectionNotification) SetPayload(payload []byte) error {
	const minPayloadSize = 3
	if len(payload) < minPayloadSize {
		return NewPayloadError(len(payload), minPayloadSize)
	}

	p.Unknown = payload[0]

	var err error
	p.Notification, _, err = readDynamicString(payload[1:])
	if err != nil {
		return err
	}

	return nil
}

type ReconnectionNotification struct {
	PacketHeader
	_    uint16 // len(Host)
	Host string
	Port uint16
}

func (p *ReconnectionNotification) Payload() ([]byte, error) {
	hostPayload, err := writeDynamicString(p.Host)
	if err != nil {
		return nil, err
	}

	var portPayload [2]byte
	binary.BigEndian.PutUint16(portPayload[:], p.Port)

	return append(hostPayload[:], portPayload[:]...), nil
}

func (p *ReconnectionNotification) SetPayload(payload []byte) error {
	const requiredPayloadSize = 4
	if len(payload) < requiredPayloadSize {
		return NewPayloadError(len(payload), requiredPayloadSize)
	}

	var n int
	var err error
	p.Host, n, err = readDynamicString(payload[:])
	if err != nil {
		return err
	}

	if len(payload) < n+2 {
		return NewPayloadError(len(payload), n+2)
	}

	p.Port = binary.BigEndian.Uint16(payload[n : n+2])

	return nil
}

type ConnectionSummaryNotification struct {
	PacketHeader
	Success bool
	Unknown uint16
}

func (p *ConnectionSummaryNotification) Payload() ([]byte, error) {
	var payload [3]byte

	if p.Success {
		payload[0] = 1
	}

	binary.BigEndian.PutUint16(payload[1:3], p.Unknown)

	return payload[:], nil
}

func (p *ConnectionSummaryNotification) SetPayload(payload []byte) error {
	const requiredPayloadSize = 3
	if len(payload) != 3 {
		return NewPayloadError(len(payload), requiredPayloadSize)
	}

	if payload[0] > 0 {
		p.Success = true
	}

	p.Unknown = binary.BigEndian.Uint16(payload[1:3])

	return nil
}

var (
	packetGenerator map[PacketTypeID]map[PacketNameID]func() Packet
)

func init() {
	packetGenerator = map[PacketTypeID]map[PacketNameID]func() Packet{
		requestID: {
			onlineCheckID:                     func() Packet { return &OnlineCheckRequest{} },
			disconnectionID:                   func() Packet { return &DisconnectionRequest{} },
			fastDataID:                        func() Packet { return &FastDataRequest{} },
			authenticationInformationHeaderID: func() Packet { return &AuthenticationInformationRequestHeader{} },
			authenticationInformationDataID:   func() Packet { return &AuthenticationInformationRequestData{} },
			authenticationInformationFooterID: func() Packet { return &AuthenticationInformationRequestFooter{} },
			tusCommonAreaAcquisitionID:        func() Packet { return &TusCommonAreaAcquisitionRequest{} },
			tusCommonAreaSettingsID:           func() Packet { return &TusCommonAreaSettingsRequest{} },
			tusCommonAreaAddID:                func() Packet { return &TusCommonAreaAddRequest{} },
			tusUserAreaWriteHeaderID:          func() Packet { return &TusUserAreaWriteRequestHeader{} },
			tusUserAreaWriteDataID:            func() Packet { return &TusUserAreaWriteRequestData{} },
			tusUserAreaWriteFooterID:          func() Packet { return &TusUserAreaWriteRequestFooter{} },
			tusUserAreaReadHeaderID:           func() Packet { return &TusUserAreaReadRequestHeader{} },
			tusUserAreaReadDataID:             func() Packet { return &TusUserAreaReadRequestData{} },
			tusUserAreaReadFooterID:           func() Packet { return &TusUserAreaReadRequestFooter{} },
		},
		responseID: {
			onlineCheckID:                     func() Packet { return &OnlineCheckResponse{} },
			disconnectionID:                   func() Packet { return &DisconnectionResponse{} },
			fastDataID:                        func() Packet { return &FastDataResponse{} },
			authenticationInformationHeaderID: func() Packet { return &AuthenticationInformationResponseHeader{} },
			authenticationInformationDataID:   func() Packet { return &AuthenticationInformationResponseData{} },
			authenticationInformationFooterID: func() Packet { return &AuthenticationInformationResponseFooter{} },
			tusCommonAreaAcquisitionID:        func() Packet { return &TusCommonAreaAcquisitionResponse{} },
			tusCommonAreaSettingsID:           func() Packet { return &TusCommonAreaSettingsResponse{} },
			tusCommonAreaAddID:                func() Packet { return &TusCommonAreaAddResponse{} },
			tusUserAreaWriteHeaderID:          func() Packet { return &TusUserAreaWriteResponseHeader{} },
			tusUserAreaWriteDataID:            func() Packet { return &TusUserAreaWriteResponseData{} },
			tusUserAreaWriteFooterID:          func() Packet { return &TusUserAreaWriteResponseFooter{} },
			tusUserAreaReadHeaderID:           func() Packet { return &TusUserAreaReadResponseHeader{} },
			tusUserAreaReadDataID:             func() Packet { return &TusUserAreaReadResponseData{} },
			tusUserAreaReadFooterID:           func() Packet { return &TusUserAreaReadResponseFooter{} },
		},
		notificationID: {
			disconnectionID:     func() Packet { return &DisconnectionNotification{} },
			reconnectionID:      func() Packet { return &ReconnectionNotification{} },
			connectionSummaryID: func() Packet { return &ConnectionSummaryNotification{} },
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

func GetPacketType(p Packet) PacketType {
	switch p.(type) {
	case *OnlineCheckRequest:
		return PacketType{onlineCheckID, requestID, noErrorID}
	case *DisconnectionRequest:
		return PacketType{disconnectionID, requestID, noErrorID}
	case *FastDataRequest:
		return PacketType{fastDataID, requestID, noErrorID}
	case *AuthenticationInformationRequestHeader:
		return PacketType{authenticationInformationHeaderID, requestID, noErrorID}
	case *AuthenticationInformationRequestData:
		return PacketType{authenticationInformationDataID, requestID, noErrorID}
	case *AuthenticationInformationRequestFooter:
		return PacketType{authenticationInformationFooterID, requestID, noErrorID}
	case *TusCommonAreaAcquisitionRequest:
		return PacketType{tusCommonAreaAcquisitionID, requestID, noErrorID}
	case *TusCommonAreaSettingsRequest:
		return PacketType{tusCommonAreaSettingsID, requestID, noErrorID}
	case *TusCommonAreaAddRequest:
		return PacketType{tusCommonAreaAddID, requestID, noErrorID}
	case *TusUserAreaWriteRequestHeader:
		return PacketType{tusUserAreaWriteHeaderID, requestID, noErrorID}
	case *TusUserAreaWriteRequestData:
		return PacketType{tusUserAreaWriteDataID, requestID, noErrorID}
	case *TusUserAreaWriteRequestFooter:
		return PacketType{tusUserAreaWriteFooterID, requestID, noErrorID}
	case *TusUserAreaReadRequestHeader:
		return PacketType{tusUserAreaReadHeaderID, requestID, noErrorID}
	case *TusUserAreaReadRequestData:
		return PacketType{tusUserAreaReadDataID, requestID, noErrorID}
	case *TusUserAreaReadRequestFooter:
		return PacketType{tusUserAreaReadFooterID, requestID, noErrorID}
	case *OnlineCheckResponse:
		return PacketType{onlineCheckID, responseID, noErrorID}
	case *DisconnectionResponse:
		return PacketType{disconnectionID, responseID, noErrorID}
	case *FastDataResponse:
		return PacketType{fastDataID, responseID, noErrorID}
	case *AuthenticationInformationResponseHeader:
		return PacketType{authenticationInformationHeaderID, responseID, noErrorID}
	case *AuthenticationInformationResponseData:
		return PacketType{authenticationInformationDataID, responseID, noErrorID}
	case *AuthenticationInformationResponseFooter:
		return PacketType{authenticationInformationFooterID, responseID, noErrorID}
	case *TusCommonAreaAcquisitionResponse:
		return PacketType{tusCommonAreaAcquisitionID, responseID, noErrorID}
	case *TusCommonAreaSettingsResponse:
		return PacketType{tusCommonAreaSettingsID, responseID, noErrorID}
	case *TusCommonAreaAddResponse:
		return PacketType{tusCommonAreaAddID, responseID, noErrorID}
	case *TusUserAreaWriteResponseHeader:
		return PacketType{tusUserAreaWriteHeaderID, responseID, noErrorID}
	case *TusUserAreaWriteResponseData:
		return PacketType{tusUserAreaWriteDataID, responseID, noErrorID}
	case *TusUserAreaWriteResponseFooter:
		return PacketType{tusUserAreaWriteFooterID, responseID, noErrorID}
	case *TusUserAreaReadResponseHeader:
		return PacketType{tusUserAreaReadHeaderID, responseID, noErrorID}
	case *TusUserAreaReadResponseData:
		return PacketType{tusUserAreaReadDataID, responseID, noErrorID}
	case *TusUserAreaReadResponseFooter:
		return PacketType{tusUserAreaReadFooterID, responseID, noErrorID}
	case *DisconnectionNotification:
		return PacketType{disconnectionID, notificationID, noErrorID}
	case *ReconnectionNotification:
		return PacketType{reconnectionID, notificationID, noErrorID}
	case *ConnectionSummaryNotification:
		return PacketType{connectionSummaryID, notificationID, noErrorID}
	default:
		return PacketType{unknownNameID, unknownTypeID, unknownErrorID}
	}
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
	return fmt.Sprintf("invalid packet type '%s', expected '%s'", e.Actual, e.Expected)
}

func NewPacketTypeError(expected interface{}, actual interface{}) PacketTypeError {
	return PacketTypeError{
		Expected: reflect.ValueOf(expected).Type().String(),
		Actual:   reflect.ValueOf(actual).Type().String(),
	}
}

func readDynamicString(data []byte) (s string, n int, err error) {
	b, n, err := readDynamicData(data[:])
	if err != nil {
		return s, n, err
	}

	return string(b), n, nil
}

func readDynamicData(data []byte) (b []byte, n int, err error) {
	const sizePrefixLength = 2
	if len(data) < sizePrefixLength {
		return b, n, NewPayloadError(len(data), sizePrefixLength)
	}

	n = int(binary.BigEndian.Uint16(data[0:2]))
	if (len(data) - sizePrefixLength) < n {
		return b, n, NewPayloadError(len(data), n+sizePrefixLength)
	}

	b = make([]byte, n)
	copy(b, data[2:n+2])
	return b, n + 2, nil
}

func writeDynamicString(s string) ([]byte, error) {
	return writeDynamicData([]byte(s))
}

func writeDynamicData(data []byte) ([]byte, error) {
	if len(data) > maxPacketLength {
		return nil, NewPayloadError(len(data), maxPacketLength)
	}

	payload := make([]byte, len(data)+2)
	binary.BigEndian.PutUint16(payload[0:2], uint16(len(data)))
	copy(payload[2:], data[:])

	return payload, nil
}
