package tools

import (
	"context"
	"encoding/json"
	"strings"

	"vcontext/internal/db"
	"vcontext/internal/mcp"
)

type GetContextParams struct {
	ID string `json:"id"`
}

func GetContextHandler(store *db.DB) mcp.Handler {
	return func(ctx context.Context, params json.RawMessage) (any, *mcp.RPCError) {
		var input GetContextParams
		if err := decodeParams(params, &input); err != nil {
			return nil, err
		}

		id := strings.TrimSpace(input.ID)
		if id == "" {
			return nil, mcp.NewError(mcp.ErrInvalidParams, "id is required")
		}

		item, err := store.GetContext(ctx, id)
		if err != nil {
			if err == db.ErrNotFound {
				return nil, mcp.NewError(-32004, "context item not found")
			}
			return nil, mcp.NewError(mcp.ErrInternal, err.Error())
		}

		return item, nil
	}
}
