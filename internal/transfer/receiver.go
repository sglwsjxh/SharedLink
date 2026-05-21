package transfer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"SharedLink/internal/protocol"
)

func Receive(ctx context.Context, remoteAddr string, progress func(receivedBytes int64, totalBytes int64)) error {
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", remoteAddr)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return err
	}

	var closeOnce sync.Once
	closeConn := func() {
		closeOnce.Do(func() {
			_ = conn.Close()
		})
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			closeConn()
		case <-done:
		}
	}()
	defer func() {
		close(done)
		closeConn()
	}()

	fileInfoHeader, fileInfoPayload, err := readPacket(ctx, conn)
	if err != nil {
		return err
	}
	if err := handleSenderError(fileInfoHeader, fileInfoPayload); err != nil {
		return err
	}
	if fileInfoHeader.Type != protocol.TypeFileInfo {
		return fmt.Errorf("expected file info packet, got %d", fileInfoHeader.Type)
	}

	fileInfo, err := protocol.UnmarshalFileInfo(fileInfoPayload)
	if err != nil {
		return err
	}
	if fileInfo.FileSize < 0 {
		return fmt.Errorf("invalid file size: %d", fileInfo.FileSize)
	}
	if fileInfo.TotalChunks != protocol.CalculateChunks(fileInfo.FileSize) {
		return fmt.Errorf("invalid chunk count: %d", fileInfo.TotalChunks)
	}

	targetPath := uniqueOutputPath(fileInfo.FileName)
	tempFile, err := os.CreateTemp(filepath.Dir(targetPath), "."+filepath.Base(targetPath)+".*.part")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	cleanup := true
	defer func() {
		_ = tempFile.Close()
		if cleanup {
			_ = os.Remove(tempPath)
		}
	}()

	var receivedBytes int64
	for expectedIndex := int32(0); expectedIndex < fileInfo.TotalChunks; expectedIndex++ {
		header, payload, err := readPacket(ctx, conn)
		if err != nil {
			return err
		}
		if err := handleSenderError(header, payload); err != nil {
			return err
		}
		if header.Type != protocol.TypeDataChunk {
			return fmt.Errorf("expected data chunk packet, got %d", header.Type)
		}

		chunk, err := protocol.UnmarshalDataChunk(payload)
		if err != nil {
			return err
		}
		if chunk.ChunkIndex != expectedIndex {
			return fmt.Errorf("unexpected chunk index: got %d, want %d", chunk.ChunkIndex, expectedIndex)
		}
		if protocol.ComputeChecksum(chunk.Data) != chunk.ChunkChecksum {
			return fmt.Errorf("checksum mismatch for chunk %d", chunk.ChunkIndex)
		}

		written, err := tempFile.Write(chunk.Data)
		if err != nil {
			return err
		}
		if written != len(chunk.Data) {
			return errors.New("short write")
		}
		receivedBytes += int64(written)
		if receivedBytes > fileInfo.FileSize {
			return fmt.Errorf("received more bytes than expected: %d > %d", receivedBytes, fileInfo.FileSize)
		}
		if progress != nil {
			progress(receivedBytes, fileInfo.FileSize)
		}
	}

	if receivedBytes != fileInfo.FileSize {
		return fmt.Errorf("received size mismatch: got %d, want %d", receivedBytes, fileInfo.FileSize)
	}

	doneHeader, donePayload, err := readPacket(ctx, conn)
	if err != nil {
		return err
	}
	if err := handleSenderError(doneHeader, donePayload); err != nil {
		return err
	}
	if doneHeader.Type != protocol.TypeTransferDone {
		return fmt.Errorf("expected transfer done packet, got %d", doneHeader.Type)
	}
	transferDone, err := protocol.UnmarshalTransferDone(donePayload)
	if err != nil {
		return err
	}

	if err := tempFile.Close(); err != nil {
		return err
	}
	checksum, err := checksumFile(tempPath)
	if err != nil {
		return err
	}
	if checksum != fileInfo.FileChecksum {
		return errors.New("file checksum mismatch")
	}
	if checksum != transferDone.FileChecksum {
		return errors.New("transfer done checksum mismatch")
	}
	if err := os.Rename(tempPath, targetPath); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func readPacket(ctx context.Context, conn net.Conn) (protocol.Header, []byte, error) {
	if err := ctx.Err(); err != nil {
		return protocol.Header{}, nil, err
	}
	header, payload, err := protocol.ReadPacket(conn)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return protocol.Header{}, nil, ctxErr
		}
		return protocol.Header{}, nil, err
	}
	return header, payload, nil
}

func handleSenderError(header protocol.Header, payload []byte) error {
	if header.Type != protocol.TypeError {
		return nil
	}
	errorPayload, err := protocol.UnmarshalErrorPayload(payload)
	if err != nil {
		return err
	}
	return errors.New(errorPayload.Message)
}

func uniqueOutputPath(fileName string) string {
	baseName := filepath.Base(fileName)
	if baseName == "." || baseName == string(filepath.Separator) || baseName == "" {
		baseName = "received_file"
	}
	baseName = strings.TrimSpace(baseName)
	if baseName == "" {
		baseName = "received_file"
	}

	dir := "."
	candidate := filepath.Join(dir, baseName)
	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		return candidate
	}

	ext := filepath.Ext(baseName)
	name := strings.TrimSuffix(baseName, ext)
	for i := 1; ; i++ {
		candidate = filepath.Join(dir, fmt.Sprintf("%s_%d%s", name, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

func checksumFile(path string) ([32]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return [32]byte{}, err
	}
	defer file.Close()
	return protocol.ComputeFileChecksum(file)
}
