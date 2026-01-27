package vcontext

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

type Client struct {
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex
	nextID uint64
}

func NewClient(r io.Reader, w io.Writer) *Client {
	return &Client{
		reader: bufio.NewReader(r),
		writer: w,
	}
}

func (c *Client) SaveContext(ctx context.Context, params SaveContextParams) (SaveContextResult, error) {
	var result SaveContextResult
	err := c.call(ctx, "tools/save_context/invoke", params, &result)
	return result, err
}

func (c *Client) SearchContext(ctx context.Context, params SearchContextParams) (SearchContextResult, error) {
	var result SearchContextResult
	err := c.call(ctx, "tools/search_context/invoke", params, &result)
	return result, err
}

func (c *Client) GetContext(ctx context.Context, params GetContextParams) (ContextItem, error) {
	var result ContextItem
	err := c.call(ctx, "tools/get_context/invoke", params, &result)
	return result, err
}

func (c *Client) call(ctx context.Context, method string, params any, out any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nextID++
	idStr := strconv.FormatUint(c.nextID, 10)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(idStr),
		Method:  method,
		Params:  params,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	if _, err := c.writer.Write(payload); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := c.reader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("connection closed")
			}
			return err
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var resp JSONRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}

		if strings.TrimSpace(string(resp.ID)) != idStr {
			continue
		}

		if resp.Error != nil {
			return resp.Error
		}
		if out == nil {
			return nil
		}
		if len(resp.Result) == 0 {
			return nil
		}
		if err := json.Unmarshal(resp.Result, out); err != nil {
			return err
		}
		return nil
	}
}
