package vnet

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Maximum message size
const (
	maxMessageSize = 1024 * 1024 // 1MB
)

// Format: [length(4 bytes)][data(length bytes)]

// ReadMessage reads a framed message from the reader
func ReadMessage(r io.Reader) ([]byte, error) {
	// Read length (4 bytes)
	var length uint32
	err := binary.Read(r, binary.LittleEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("read message length error: %w", err)
	}
	// Check length to prevent DoS
	if length == 0 {
		return nil, fmt.Errorf("message length is 0")
	}

	if length > maxMessageSize {
		return nil, fmt.Errorf("message too large: %d > %d", length, maxMessageSize)
	}
	// Read message data
	data := make([]byte, length)
	_, err = io.ReadFull(r, data)
	if err != nil {
		return nil, fmt.Errorf("read message data error: %w", err)
	}

	return data, nil
}

// WriteMessage writes a framed message to the writer
func WriteMessage(w io.Writer, data []byte) error {
	// Get data length
	length := uint32(len(data))
	if length == 0 {
		return fmt.Errorf("message data length is 0")
	}
	if length > maxMessageSize {
		return fmt.Errorf("message too large: %d > %d", length, maxMessageSize)
	}
	// Write length
	err := binary.Write(w, binary.LittleEndian, length)
	if err != nil {
		return fmt.Errorf("write message length error: %w", err)
	}

	// Write message data
	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("write message data error: %w", err)
	}

	return nil
}
