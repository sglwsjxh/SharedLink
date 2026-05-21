package transfer

import (
	"bufio"
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"SharedLink/internal/protocol"
)

func Send(ctx context.Context, addr string, filePath string, progress func(sentBytes int64, totalBytes int64)) error {
	if addr == "" {
		addr = net.JoinHostPort("0.0.0.0", strconv.Itoa(protocol.DefaultPort))
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

	fileChecksum, err := protocol.ComputeFileChecksum(file)
	if err != nil {
		return err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
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
		FileName:     filepath.Base(filePath),
		FileSize:     fileSize,
		TotalChunks:  totalChunks,
		FileChecksum: fileChecksum,
	})
	if err := writePacket(ctx, conn, protocol.TypeFileInfo, fileInfoPayload); err != nil {
		return err
	}

	reader := bufio.NewReaderSize(file, protocol.ChunkSize)
	sentBytes := int64(0)
	for chunkIndex := range totalChunks {
		if err := ctx.Err(); err != nil {
			return err
		}

		remaining := fileSize - sentBytes
		chunkSize := min(int64(protocol.ChunkSize), remaining)

		data := make([]byte, chunkSize)
		if _, err := io.ReadFull(reader, data); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}

		payload := protocol.MarshalDataChunk(protocol.DataChunk{
			ChunkIndex:    chunkIndex,
			Data:          data,
			ChunkChecksum: protocol.ComputeChecksum(data),
		})
		if err := writePacket(ctx, conn, protocol.TypeDataChunk, payload); err != nil {
			return err
		}

		sentBytes += int64(len(data))
		if progress != nil {
			progress(sentBytes, fileSize)
		}
	}

	donePayload := protocol.MarshalTransferDone(protocol.TransferDone{FileChecksum: fileChecksum})
	if err := writePacket(ctx, conn, protocol.TypeTransferDone, donePayload); err != nil {
		return err
	}

	if _, err := io.Copy(io.Discard, conn); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
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
