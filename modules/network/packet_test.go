package network

import (
	"bytes"
	"testing"
)

func TestBooleanPacket(t *testing.T) {
	in := BooleanPacket{Value: true}
	out := BooleanPacket{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if in.Value != out.Value {
		t.Error("Value")
	}
}

func TestPropertyPacket(t *testing.T) {
	properties := []Property{
		{Index: 0, Value1: 1, Value2: 2},
		{Index: 1, Value1: 3, Value2: 4},
	}

	in := PropertyPacket{Properties: properties}
	out := PropertyPacket{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if len(in.Properties) != len(out.Properties) {
		t.Error("Properties")
		return
	}

	for i := 0; i < len(in.Properties); i++ {
		inProp := in.Properties[i]
		outProp := out.Properties[i]

		if inProp.Index != outProp.Index {
			t.Error("Index")
		}

		if inProp.Value1 != outProp.Value1 {
			t.Error("Value1")
		}

		if inProp.Value2 != outProp.Value2 {
			t.Error("Value2")
		}

	}
}

func TestDataChunkReferencePacket(t *testing.T) {
	in := DataChunkReferencePacket{ChunkOffset: 10, ChunkLength: 20}
	out := DataChunkReferencePacket{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if in.ChunkOffset != out.ChunkOffset {
		t.Error("ChunkOffset")
	}

	if in.ChunkLength != out.ChunkLength {
		t.Error("ChunkLength")
	}
}

func TestDataChunkPacket(t *testing.T) {
	var chunk [5]byte
	for i := 0; i < len(chunk); i++ {
		chunk[i] = byte(i)
	}

	in := DataChunkPacket{ChunkOffset: 10, ChunkData: chunk[:]}
	out := DataChunkPacket{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if in.ChunkOffset != out.ChunkOffset {
		t.Error("ChunkOffset")
	}

	if !bytes.Equal(in.ChunkData[:], out.ChunkData[:]) {
		t.Error("ChunkData")
	}
}

//func TestOnlineCheckRequest(t *testing.T) {
//See: EmptyPacket

//func TestDisconnectionRequest(t *testing.T) {
//See: BooleanPacket

//func TestFastDataRequest(t *testing.T) {
//See: EmptyPacket

func TestAuthenticationInformationRequestHeader(t *testing.T) {
	in := AuthenticationInformationRequestHeader{DataLength: 1024, Unknown: 2}
	out := AuthenticationInformationRequestHeader{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if in.DataLength != out.DataLength {
		t.Error("DataLength")
	}
}

//func TestAuthenticationInformationRequestData(t *testing.T) {
//See: DataChunkPacket

//func TestAuthenticationInformationRequestFooter(t *testing.T) {
//See: EmptyPacket

func TestTusCommonAreaAcquisitionRequest(t *testing.T) {
	indices := []byte{
		0x01,
		0x02,
		0x03,
	}

	in := TusCommonAreaAcquisitionRequest{PropertyIndices: indices}
	out := TusCommonAreaAcquisitionRequest{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if len(in.PropertyIndices) != len(out.PropertyIndices) {
		t.Error("PropertyIndices")
		return
	}

	for i := 0; i < len(in.PropertyIndices); i++ {
		inIndex := in.PropertyIndices[i]
		outIndex := out.PropertyIndices[i]

		if inIndex != outIndex {
			t.Error("PropertyIndex")
		}
	}
}

//func TestTusCommonAreaSettingsRequest(t *testing.T) {
//See: PropertyPacket

//func TestTusCommonAreaAddRequest(t *testing.T) {
//See: PropertyPacket

func TestTusUserAreaWriteRequestHeader(t *testing.T) {
	in := TusUserAreaWriteRequestHeader{DataLength: 1024, User: "ABCDEFGH"}
	out := TusUserAreaWriteRequestHeader{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if in.DataLength != out.DataLength {
		t.Error("DataLength")
	}

	if in.User != out.User {
		t.Error("User")
	}
}

//func TestTusUserAreaWriteRequestData(t *testing.T) {
//See: DataChunkPacket

func TestTusUserAreaWriteRequestFooter(t *testing.T) {
	in := TusUserAreaWriteRequestFooter{}
	out := TusUserAreaWriteRequestFooter{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestTusUserAreaReadRequestHeader(t *testing.T) {
	in := TusUserAreaReadRequestHeader{User: "ABCDEFGH"}
	out := TusUserAreaReadRequestHeader{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if in.User != out.User {
		t.Error("User")
	}
}

//func TestTusUserAreaReadRequestData(t *testing.T) {
//See: DataChunkReferencePacket

func TestTusUserAreaReadRequestFooter(t *testing.T) {
	in := TusUserAreaReadRequestFooter{}
	out := TusUserAreaReadRequestFooter{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestOnlineCheckResponse(t *testing.T) {
	in := OnlineCheckResponse{}
	out := OnlineCheckResponse{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}
}

//func TestDisconnectionResponse(t *testing.T) {
//See: BooleanPacket

func TestFastDataResponse(t *testing.T) {
	in := FastDataResponse{User: "132456"}
	out := FastDataResponse{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if in.User != out.User {
		t.Error("User")
	}

	if in.Unknown1 != out.Unknown1 {
		t.Error("Unknown1")
	}

	if in.Unknown2 != out.Unknown2 {
		t.Error("Unknown2")
	}
}

func TestAuthenticationInformationResponseHeader(t *testing.T) {
	in := AuthenticationInformationResponseHeader{ChunkLength: 512}
	out := AuthenticationInformationResponseHeader{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if in.ChunkLength != out.ChunkLength {
		t.Error("ChunkLength")
	}
}

//func TestAuthenticationInformationResponseData(t *testing.T) {
//See: DataChunkReferencePacket

//func TestAuthenticationInformationResponseFooter(t *testing.T) {
//See: BooleanPacket

//func TestTusCommonAreaAcquisitionResponse(t *testing.T) {
//See: PropertyPacket

//func TestTusCommonAreaSettingsResponse(t *testing.T) {
//See: PropertyPacket

//func TestTusCommonAreaAddResponse(t *testing.T) {
//See: PropertyPacket

func TestTusUserAreaWriteResponseHeader(t *testing.T) {
	in := TusUserAreaWriteResponseHeader{ChunkLength: 64}
	out := TusUserAreaWriteResponseHeader{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if in.ChunkLength != out.ChunkLength {
		t.Error("ChunkLength")
	}
}

//func TestTusUserAreaWriteResponseData(t *testing.T) {
//See: DataChunkReferencePacket

func TestTusUserAreaWriteResponseFooter(t *testing.T) {
	in := TusUserAreaWriteResponseFooter{}
	out := TusUserAreaWriteResponseFooter{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestTusUserAreaReadResponseHeader(t *testing.T) {
	in := TusUserAreaReadResponseHeader{DataLength: 256}
	out := TusUserAreaReadResponseHeader{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if in.DataLength != out.DataLength {
		t.Error("DataLength")
	}
}

//func TestTusUserAreaReadResponseData(t *testing.T) {
//See: DataChunkPacket

func TestTusUserAreaReadResponseFooter(t *testing.T) {
	in := TusUserAreaReadResponseFooter{}
	out := TusUserAreaReadResponseFooter{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestDisconnectionNotification(t *testing.T) {
	in := DisconnectionNotification{Unknown: 0xAA, Notification: "TestNotification"}
	out := DisconnectionNotification{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if in.Unknown != out.Unknown {
		t.Error("Unknown")
	}

	if in.Notification != out.Notification {
		t.Error("Notification")
	}
}

func TestReconnectionNotification(t *testing.T) {
	in := ReconnectionNotification{Host: "localhost", Port: 11111}
	out := ReconnectionNotification{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if in.Host != out.Host {
		t.Error("Host")
	}

	if in.Port != out.Port {
		t.Error("Port")
	}
}

func TestConnectionSummaryNotification(t *testing.T) {
	in := ConnectionSummaryNotification{Unknown: 10, Success: true}
	out := ConnectionSummaryNotification{}

	err := roundtrip(&in, &out)
	if err != nil {
		t.Error(err)
		return
	}

	if in.Unknown != out.Unknown {
		t.Error("Unknown")
	}

	if in.Success != out.Success {
		t.Error("Success")
	}
}

func roundtrip(in Packet, out Packet) error {
	payload, err := in.Payload()
	if err != nil {
		return err
	}

	err = out.SetPayload(payload)
	if err != nil {
		return err
	}

	return nil
}
