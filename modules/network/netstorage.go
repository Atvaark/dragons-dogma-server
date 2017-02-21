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

const userAreaType = uint32(0x12122700)

var userAreaKey = []byte("nokupak amugod uznogarod")

func ReadUserArea(data []byte) (*UserArea, error) {
	if len(data) == 0 {
		return nil, nil
	}

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

func WriteUserArea(area *UserArea) ([]byte, error) {
	if area == nil {
		return make([]byte, 0), nil
	}

	data, err := serializeUserArea(area)
	if err != nil {
		return nil, err
	}

	data, err = compressUserArea(data)
	if err != nil {
		return nil, err
	}

	data, err = encryptUserArea(data)
	if err != nil {
		return nil, err
	}

	data, err = padUserArea(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func padUserArea(data []byte) ([]byte, error) {
	const targetSize = 2048
	if len(data) > targetSize {
		return nil, errors.New("user area exceeds max size")
	}

	if len(data) == targetSize {
		return data, nil
	}

	padded := make([]byte, targetSize)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = 0xDD
	}

	return padded, nil
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

type encryptedHeader struct {
	Type   uint32 // 0x12122700, 0x11090800
	Length int
}

func encryptUserArea(data []byte) ([]byte, error) {
	var h encryptedHeader
	h.Type = userAreaType
	h.Length = len(data)

	dataEncrypted, err := encrypt(userAreaKey, data)
	if err != nil {
		return nil, err
	}

	encrypted := make([]byte, 8+len(dataEncrypted))
	binary.LittleEndian.PutUint32(encrypted[0:4], h.Type)
	binary.LittleEndian.PutUint32(encrypted[4:8], uint32(h.Length))
	copy(encrypted[8:], dataEncrypted)

	return encrypted, nil
}

func encrypt(key, data []byte) ([]byte, error) {
	ci, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := ci.BlockSize()
	pad := requiredPadding(len(data), blockSize)
	dataEncrypted := make([]byte, len(data)+pad)
	copy(dataEncrypted, data)

	for i := 0; i < len(dataEncrypted); i += blockSize {
		swapUint32Endianess(dataEncrypted[i : i+8])
		ci.Encrypt(dataEncrypted[i:i+blockSize], dataEncrypted[i:i+blockSize])
		swapUint32Endianess(dataEncrypted[i : i+8])
	}

	return dataEncrypted[:], nil
}

func decryptUserArea(data []byte) ([]byte, error) {
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

	dataDecrypted, err := decrypt(userAreaKey, data[8:], h.Length)
	if err != nil {
		return nil, err
	}

	return dataDecrypted[:], nil
}

func decrypt(key, data []byte, n int) ([]byte, error) {
	ci, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := ci.BlockSize()
	pad := requiredPadding(len(data), blockSize)
	dataDecrypted := make([]byte, len(data)+pad)
	copy(dataDecrypted, data)

	for i := 0; i < len(dataDecrypted); i += blockSize {
		swapUint32Endianess(dataDecrypted[i : i+8])
		ci.Decrypt(dataDecrypted[i:i+blockSize], dataDecrypted[i:i+blockSize])
		swapUint32Endianess(dataDecrypted[i : i+8])
	}

	return dataDecrypted[:n], nil
}

type compressedUserAreaHeader struct {
	Type               uint32 // 0x12122700, 0x11090800
	Length             int
	LengthDecompressed int
	SHA1Hash           [20]byte
}

func compressUserArea(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	r := bytes.NewReader(data)
	_, err := io.Copy(zw, r)
	if err != nil {
		return nil, err
	}
	err = zw.Close()
	if err != nil {
		return nil, err
	}

	dataCompressed := buf.Bytes()

	var h compressedUserAreaHeader
	h.Type = userAreaType
	h.Length = len(dataCompressed)
	h.LengthDecompressed = len(data)

	digest := sha1.New()
	_, err = digest.Write(dataCompressed)
	if err != nil {
		return nil, err
	}
	copy(h.SHA1Hash[:], digest.Sum(nil))

	compressed := make([]byte, 32+len(dataCompressed))
	binary.LittleEndian.PutUint32(compressed[0:4], h.Type)
	binary.LittleEndian.PutUint32(compressed[4:8], uint32(h.Length))
	binary.LittleEndian.PutUint32(compressed[8:12], uint32(h.LengthDecompressed))
	copy(compressed[12:32], h.SHA1Hash[:])
	swapUint32Endianess(compressed[12:32])
	copy(compressed[32:], dataCompressed)

	return compressed, nil
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

	dataCompressed := data[32:]

	digest := sha1.New()
	_, err := digest.Write(dataCompressed)
	if err != nil {
		return nil, err
	}
	hash := digest.Sum(nil)

	if !bytes.Equal(h.SHA1Hash[:], hash) {
		return nil, errors.New("hash mismatch")
	}

	r := bytes.NewReader(dataCompressed)
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

const userAreaItemLen = 4
const userAreaSlotsLen = SlotsMax * (13 + ItemsMax*userAreaItemLen)
const userAreaHeaderLen = 8
const userAreaLen = userAreaHeaderLen + userAreaSlotsLen

func serializeUserArea(area *UserArea) ([]byte, error) {
	data := make([]byte, userAreaLen)

	binary.BigEndian.PutUint32(data[0:4], area.Unknown)
	binary.BigEndian.PutUint32(data[4:8], area.UnknownCount)

	var offset int = 8
	for i := 0; i < len(area.Slots); i++ {
		slot := &area.Slots[i]

		data[offset] = slot.Unknown
		offset += 1

		for j := 0; j < len(slot.Items); j++ {
			binary.BigEndian.PutUint32(data[offset:offset+4], uint32(slot.Items[j]))
			offset += 4
		}

		binary.BigEndian.PutUint32(data[offset:offset+4], slot.ItemsCount)
		binary.BigEndian.PutUint64(data[offset+4:offset+12], slot.User)
		offset += 12
	}

	return data, nil
}

func parseUserArea(data []byte) (*UserArea, error) {
	if len(data) < userAreaLen {
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
