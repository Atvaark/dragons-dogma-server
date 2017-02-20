package network

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/blowfish"
)

const SlotsMax = 100

type UserArea struct {
	Unknown      uint32 // 0
	UnknownCount uint32
	Slots        [SlotsMax]UserAreaSlot
}

const ItemsMax = 10

type UserAreaSlot struct {
	Unknown    byte // 0=used, 1=unused
	Items      [ItemsMax]UserAreaItem
	ItemsCount uint32
	User       uint64
}

type UserAreaItem uint32

func ReadUserArea(data []byte) (*UserArea, error) {
	data, err := decryptUserArea(data)
	if err != nil {
		return nil, err
	}

	data, err = decompressUserArea(data)
	if err != nil {
		return nil, err
	}

	area, err := parseUserArea(data)
	if err != nil {
		return nil, err
	}

	return area, nil
}

func swapUint32Endianess(b []byte) {
	for i := 0; i < len(b); i += 4 {
		b[i], b[i+1], b[i+2], b[i+3] = b[i+3], b[i+2], b[i+1], b[i]
	}
}

func requiredPadding(dataLength int, blockSize int) int {
	mod := dataLength % blockSize
	if mod == 0 {
		return 0
	}

	return blockSize - mod
}

var userAreaKey = []byte("nokupak amugod uznogarod")

func decryptUserArea(data []byte) ([]byte, error) {
	type encryptedHeader struct {
		Type   uint32 // 0x12122700, 0x11090800
		Length int
	}

	if len(data) == 0 {
		return make([]byte, 0), nil
	}

	if len(data) < 8 {
		return nil, errors.New("insufficient data to decrypt")
	}

	var h encryptedHeader
	h.Type = binary.LittleEndian.Uint32(data[0:4])
	h.Length = int(binary.LittleEndian.Uint32(data[4:8]))

	if len(data)-8 < h.Length {
		return nil, errors.New("insufficient data to decrypt")
	}

	ci, err := blowfish.NewCipher(userAreaKey)
	if err != nil {
		return nil, err
	}

	blockSize := ci.BlockSize()
	pad := requiredPadding(h.Length, blockSize)
	buf := make([]byte, h.Length+pad)
	copy(buf, data[8:8+h.Length])

	for i := 0; i < len(buf); i += blockSize {
		swapUint32Endianess(buf[i : i+8])
		ci.Decrypt(buf[i:i+blockSize], buf[i:i+blockSize])
		swapUint32Endianess(buf[i : i+8])
	}

	return buf[:h.Length], nil
}

type compressedUserAreaHeader struct {
	Type               uint32 // 0x12122700, 0x11090800
	Length             int
	LengthDecompressed int
	SHA1Hash           [20]byte
}

func decompressUserArea(data []byte) ([]byte, error) {
	if len(data) < 32 {
		return nil, errors.New("insufficient data to decompress")
	}

	var h compressedUserAreaHeader
	h.Type = binary.LittleEndian.Uint32(data[0:4])
	h.Length = int(binary.LittleEndian.Uint32(data[4:8]))
	h.LengthDecompressed = int(binary.LittleEndian.Uint32(data[8:12]))
	copy(h.SHA1Hash[:], data[12:32])
	swapUint32Endianess(h.SHA1Hash[:])

	if len(data)-32 < h.Length {
		return nil, errors.New("insufficient data to decompress")
	}

	compressed := data[32:]

	digest := sha1.New()
	_, err := digest.Write(compressed)
	if err != nil {
		return nil, err
	}
	hash := digest.Sum(nil)

	if !bytes.Equal(h.SHA1Hash[:], hash) {
		return nil, errors.New("hash mismatch")
	}

	r := bytes.NewReader(compressed)
	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, h.LengthDecompressed))

	_, err = io.Copy(buf, zr)
	if err != nil {
		return nil, err
	}

	err = zr.Close()
	if err != nil {
		return nil, err
	}

	if buf.Len() != h.LengthDecompressed {
		return nil, errors.New("invalid size")
	}

	return buf.Bytes(), nil
}

func parseUserArea(data []byte) (*UserArea, error) {
	const itemLen = 4
	const slotsLen = SlotsMax * (13 + ItemsMax*itemLen)
	const headerLen = 8
	if len(data) < headerLen+slotsLen {
		return nil, errors.New("insufficient data to decompress")
	}

	var area UserArea
	area.Unknown = binary.BigEndian.Uint32(data[0:4])
	area.UnknownCount = binary.BigEndian.Uint32(data[4:8])

	var offset int = 8
	for i := 0; i < len(area.Slots); i++ {
		slot := &area.Slots[i]

		slot.Unknown = data[offset]
		offset += 1

		for j := 0; j < len(slot.Items); j++ {
			slot.Items[j] = UserAreaItem(binary.BigEndian.Uint32(data[offset : offset+4]))
			offset += 4
		}

		slot.ItemsCount = binary.BigEndian.Uint32(data[offset : offset+4])
		slot.User = binary.BigEndian.Uint64(data[offset+4 : offset+12])
		offset += 12
	}

	return &area, nil
}
