package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	Magic          = "SHLN"
	DefaultPort    = 53349
	DefaultPortStr = "53349"
	ChunkSize      = 4 * 1024 * 1024
	ServiceType    = "_sharedlink._tcp"
	HeaderSize     = 9
)

type PacketType uint8

const (
	TypeFileInfo     PacketType = 0
	TypeDataChunk    PacketType = 1
	TypeTransferDone PacketType = 2
	TypeError        PacketType = 3
)

type Header struct {
	Magic      [4]byte
	Type       PacketType
	PayloadLen uint32
}

func NewHeader(t PacketType, payloadLen uint32) Header {
	var h Header
	copy(h.Magic[:], Magic)
	h.Type = t
	h.PayloadLen = payloadLen
	return h
}

func (h *Header) Valid() bool {
	return string(h.Magic[:]) == Magic
}

func MarshalHeader(h Header) ([]byte, error) {
	buf := make([]byte, HeaderSize)
	copy(buf[0:4], h.Magic[:])
	buf[4] = byte(h.Type)
	binary.BigEndian.PutUint32(buf[5:9], h.PayloadLen)
	return buf, nil
}

func UnmarshalHeader(buf []byte) (Header, error) {
	if len(buf) < HeaderSize {
		return Header{}, fmt.Errorf("header too short: %d bytes", len(buf))
	}
	var h Header
	copy(h.Magic[:], buf[0:4])
	h.Type = PacketType(buf[4])
	h.PayloadLen = binary.BigEndian.Uint32(buf[5:9])
	if !h.Valid() {
		return Header{}, fmt.Errorf("invalid magic: %q", string(h.Magic[:]))
	}
	if h.PayloadLen > 10*1024*1024 {
		return Header{}, fmt.Errorf("payload too large: %d bytes", h.PayloadLen)
	}
	return h, nil
}

type FileInfo struct {
	FileName     string
	FileSize     int64
	TotalChunks  int32
	FileChecksum [32]byte
}

func MarshalFileInfo(fi FileInfo) []byte {
	nameBytes := []byte(fi.FileName)
	payloadLen := 2 + len(nameBytes) + 8 + 4 + 32
	buf := make([]byte, payloadLen)
	pos := 0
	binary.BigEndian.PutUint16(buf[pos:], uint16(len(nameBytes)))
	pos += 2
	copy(buf[pos:], nameBytes)
	pos += len(nameBytes)
	binary.BigEndian.PutUint64(buf[pos:], uint64(fi.FileSize))
	pos += 8
	binary.BigEndian.PutUint32(buf[pos:], uint32(fi.TotalChunks))
	pos += 4
	copy(buf[pos:], fi.FileChecksum[:])
	return buf
}

func UnmarshalFileInfo(data []byte) (FileInfo, error) {
	if len(data) < 2 {
		return FileInfo{}, fmt.Errorf("fileinfo data too short")
	}
	pos := 0
	nameLen := int(binary.BigEndian.Uint16(data[pos:]))
	pos += 2
	if pos+nameLen > len(data) {
		return FileInfo{}, fmt.Errorf("fileinfo filename truncated")
	}
	fileName := string(data[pos : pos+nameLen])
	pos += nameLen
	if pos+12 > len(data) {
		return FileInfo{}, fmt.Errorf("fileinfo size/chunks truncated")
	}
	fileSize := int64(binary.BigEndian.Uint64(data[pos:]))
	pos += 8
	totalChunks := int32(binary.BigEndian.Uint32(data[pos:]))
	pos += 4
	if pos+32 > len(data) {
		return FileInfo{}, fmt.Errorf("fileinfo checksum truncated")
	}
	var checksum [32]byte
	copy(checksum[:], data[pos:pos+32])
	return FileInfo{
		FileName:     fileName,
		FileSize:     fileSize,
		TotalChunks:  totalChunks,
		FileChecksum: checksum,
	}, nil
}

type DataChunk struct {
	ChunkIndex    int32
	Data          []byte
	ChunkChecksum [32]byte
}

func MarshalDataChunk(dc DataChunk) []byte {
	payloadLen := 4 + 4 + len(dc.Data) + 32
	buf := make([]byte, payloadLen)
	pos := 0
	binary.BigEndian.PutUint32(buf[pos:], uint32(dc.ChunkIndex))
	pos += 4
	binary.BigEndian.PutUint32(buf[pos:], uint32(len(dc.Data)))
	pos += 4
	copy(buf[pos:], dc.Data)
	pos += len(dc.Data)
	copy(buf[pos:], dc.ChunkChecksum[:])
	return buf
}

func UnmarshalDataChunk(data []byte) (DataChunk, error) {
	if len(data) < 8 {
		return DataChunk{}, fmt.Errorf("datachunk header truncated")
	}
	pos := 0
	chunkIndex := int32(binary.BigEndian.Uint32(data[pos:]))
	pos += 4
	dataLen := int(binary.BigEndian.Uint32(data[pos:]))
	pos += 4
	if pos+dataLen+32 > len(data) {
		return DataChunk{}, fmt.Errorf("datachunk data truncated: need %d bytes, have %d", dataLen+32, len(data)-pos)
	}
	chunkData := make([]byte, dataLen)
	copy(chunkData, data[pos:pos+dataLen])
	pos += dataLen
	var checksum [32]byte
	copy(checksum[:], data[pos:pos+32])
	return DataChunk{
		ChunkIndex:    chunkIndex,
		Data:          chunkData,
		ChunkChecksum: checksum,
	}, nil
}

type TransferDone struct {
	FileChecksum [32]byte
}

func MarshalTransferDone(td TransferDone) []byte {
	buf := make([]byte, 32)
	copy(buf, td.FileChecksum[:])
	return buf
}

func UnmarshalTransferDone(data []byte) (TransferDone, error) {
	if len(data) < 32 {
		return TransferDone{}, fmt.Errorf("transferdone data truncated")
	}
	var td TransferDone
	copy(td.FileChecksum[:], data[:32])
	return td, nil
}

type ErrorPayload struct {
	Message string
}

func MarshalErrorPayload(ep ErrorPayload) []byte {
	msgBytes := []byte(ep.Message)
	buf := make([]byte, 2+len(msgBytes))
	binary.BigEndian.PutUint16(buf[0:2], uint16(len(msgBytes)))
	copy(buf[2:], msgBytes)
	return buf
}

func UnmarshalErrorPayload(data []byte) (ErrorPayload, error) {
	if len(data) < 2 {
		return ErrorPayload{}, fmt.Errorf("error payload too short")
	}
	msgLen := int(binary.BigEndian.Uint16(data[0:2]))
	if 2+msgLen > len(data) {
		return ErrorPayload{}, fmt.Errorf("error message truncated")
	}
	return ErrorPayload{Message: string(data[2 : 2+msgLen])}, nil
}

func WritePacket(w io.Writer, h Header, payload []byte) error {
	headerBytes, err := MarshalHeader(h)
	if err != nil {
		return err
	}
	full := make([]byte, len(headerBytes)+len(payload))
	copy(full, headerBytes)
	copy(full[len(headerBytes):], payload)
	_, err = w.Write(full)
	return err
}

func ReadPacket(r io.Reader) (Header, []byte, error) {
	headerBuf := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r, headerBuf); err != nil {
		return Header{}, nil, fmt.Errorf("read header: %w", err)
	}
	h, err := UnmarshalHeader(headerBuf)
	if err != nil {
		return Header{}, nil, fmt.Errorf("unmarshal header: %w", err)
	}
	payload := make([]byte, h.PayloadLen)
	if h.PayloadLen > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			return Header{}, nil, fmt.Errorf("read payload: %w", err)
		}
	}
	return h, payload, nil
}

func CalculateChunks(fileSize int64) int32 {
	chunks := fileSize / int64(ChunkSize)
	if fileSize%int64(ChunkSize) != 0 {
		chunks++
	}
	return int32(chunks)
}
