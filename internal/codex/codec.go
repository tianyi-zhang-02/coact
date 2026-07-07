// Package codex is a pure-Go client for the Codex app-server (JSON-RPC 2.0).
// CoAct drives Codex as a stdio executor: start/steer turns and read the
// streamed agent output. Grounded in the documented app-server protocol
// (turn/start, turn/steer with expectedTurnId, item/agentMessage/delta,
// turn/started/completed).
package codex

import (
	"bufio"
	"bytes"
	"io"
)

// Codec frames JSON-RPC messages on a byte stream. Newline-delimited is the
// default; the exact stdio framing of `codex app-server` is confirmed against a
// real codex at integration time, and the interface keeps that swappable
// (e.g. LSP Content-Length) without touching client logic.
type Codec interface {
	ReadMessage(r *bufio.Reader) ([]byte, error)
	WriteMessage(w io.Writer, msg []byte) error
}

// NewlineCodec: one JSON message per line.
type NewlineCodec struct{}

// ReadMessage returns the next non-blank line, or an error (io.EOF at end).
func (NewlineCodec) ReadMessage(r *bufio.Reader) ([]byte, error) {
	for {
		line, err := r.ReadBytes('\n')
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) > 0 {
			return trimmed, err
		}
		if err != nil {
			return nil, err
		}
	}
}

// WriteMessage writes one JSON message followed by a newline.
func (NewlineCodec) WriteMessage(w io.Writer, msg []byte) error {
	out := make([]byte, 0, len(msg)+1)
	out = append(out, msg...)
	out = append(out, '\n')
	_, err := w.Write(out)
	return err
}
