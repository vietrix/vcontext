package tools

import (
	"encoding/json"
	"strings"

	"vcontext/internal/mcp"
)

func decodeParams(raw json.RawMessage, target any) *mcp.RPCError {
	trimmed := strings.TrimSpace(string(raw))
	if len(raw) == 0 || trimmed == "" || trimmed == "null" {
		return mcp.NewError(mcp.ErrInvalidParams, "params are required")
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return mcp.NewError(mcp.ErrInvalidParams, "invalid params")
	}
	return nil
}
