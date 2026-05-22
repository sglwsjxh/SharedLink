package transfer

import (
	"context"
	"crypto/sha256"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"SharedLink/internal/protocol"
)

func Send(ctx context.Context, addr string, filePath string, progress func(sentBytes int64, totalBytes int64)) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if addr == "" {
		addr = net.JoinHostPort("0.0.0.0", strconv.Itoa(protocol.DefaultPort))
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	acceptDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			listener.Close()
		case <-acceptDone:
		}
	}()

	conn, err := listener.Accept()
	close(acceptDone)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	}
	defer conn.Close()

	connDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-connDone:
		}
	}()
	defer close(connDone)

	fileSize := info.Size()
	totalChunks := protocol.CalculateChunks(fileSize)
	fileInfoPayload := protocol.MarshalFileInfo(protocol.FileInfo{
		FileName:    filepath.Base(filePath),
		FileSize:    fileSize,
		TotalChunks: totalChunks,
		// FileInfo checksum is zeroed out — final integrity check
		// uses the TransferDone packet with incrementally computed hash.
		FileChecksum: [32]byte{},
	})
	if err := writePacket(ctx, conn, protocol.TypeFileInfo, fileInfoPayload); err != nil {
		return err
	}

	// Reusable buffer — allocated once to avoid per-chunk GC pressure
	chunkBuf := make([]byte, protocol.ChunkSize)
	fileHash := sha256.New()

	sentBytes := int64(0)
	for chunkIndex := range totalChunks {
		if err := ctx.Err(); err != nil {
			return err
		}

		remaining := fileSize - sentBytes
		chunkSize := min(int64(protocol.ChunkSize), remaining)

		data := chunkBuf[:chunkSize]
		if _, err := io.ReadFull(file, data); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		fileHash.Write(data)

		payload := protocol.MarshalDataChunk(protocol.DataChunk{
			ChunkIndex: chunkIndex,
			Data:       data,
			// Per-chunk SHA256 removed for performance;
			// end-to-end incremental hash + final checksum provides integrity.
			ChunkChecksum: [32]byte{},
		})
		if err := writePacket(ctx, conn, protocol.TypeDataChunk, payload); err != nil {
			return err
		}

		sentBytes += int64(len(data))
		if progress != nil {
			progress(sentBytes, fileSize)
		}
	}

	var transferChecksum [32]byte
	copy(transferChecksum[:], fileHash.Sum(nil))
	donePayload := protocol.MarshalTransferDone(protocol.TransferDone{FileChecksum: transferChecksum})
	if err := writePacket(ctx, conn, protocol.TypeTransferDone, donePayload); err != nil {
		return err
	}

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if _, err := io.Copy(io.Discard, conn); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil
		}
		return err
	}

	return nil
}

func writePacket(ctx context.Context, w io.Writer, packetType protocol.PacketType, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := protocol.WritePacket(w, protocol.NewHeader(packetType, uint32(len(payload))), payload); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	}
	return nil
}
