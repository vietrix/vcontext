package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
)

const maxMessageBytes = 8 * 1024 * 1024

type Handler func(ctx context.Context, params json.RawMessage) (any, *RPCError)

type Server struct {
	handlers map[string]Handler
	logger   *log.Logger
}

func NewServer(logger *log.Logger) *Server {
	return &Server{
		handlers: make(map[string]Handler),
		logger:   logger,
	}
}

func (s *Server) Register(method string, handler Handler) {
	s.handlers[method] = handler
}

func (s *Server) Serve(ctx context.Context, r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), maxMessageBytes)
	encoder := json.NewEncoder(w)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		resp := s.handleLine(ctx, line)
		if resp == nil {
			continue
		}

		if err := encoder.Encode(resp); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (s *Server) handleLine(ctx context.Context, line []byte) *JSONRPCResponse {
	var req JSONRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      json.RawMessage("null"),
			Error:   NewError(ErrParse, "parse error"),
		}
	}

	if req.JSONRPC != "2.0" || req.Method == "" {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      ensureID(req.ID),
			Error:   NewError(ErrInvalidRequest, "invalid request"),
		}
	}

	handler, ok := s.handlers[req.Method]
	if !ok {
		if len(req.ID) == 0 {
			return nil
		}
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   NewError(ErrMethodNotFound, "method not found"),
		}
	}

	result, rpcErr := handler(ctx, req.Params)
	if len(req.ID) == 0 {
		return nil
	}

	if rpcErr != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   rpcErr,
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func ensureID(id json.RawMessage) json.RawMessage {
	if len(id) == 0 {
		return json.RawMessage("null")
	}
	return id
}

func (s *Server) logError(err error, message string) {
	if s.logger == nil || err == nil {
		return
	}
	if errors.Is(err, context.Canceled) {
		return
	}
	s.logger.Printf("%s: %v", message, err)
}
