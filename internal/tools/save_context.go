package tools

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"vcontext/internal/db"
	"vcontext/internal/mcp"
)

const defaultImportance = 3

type SaveContextParams struct {
	Source     *string   `json:"source"`
	ThreadID   *string   `json:"thread_id"`
	Role       *string   `json:"role"`
	Title      *string   `json:"title"`
	Content    string    `json:"content"`
	Tags       *[]string `json:"tags"`
	Importance *int      `json:"importance"`
}

type SaveContextResult struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"created_at"`
}

func SaveContextHandler(store *db.DB) mcp.Handler {
	return func(ctx context.Context, params json.RawMessage) (any, *mcp.RPCError) {
		var input SaveContextParams
		if err := decodeParams(params, &input); err != nil {
			return nil, err
		}

		if strings.TrimSpace(input.Content) == "" {
			return nil, mcp.NewError(mcp.ErrInvalidParams, "content is required")
		}

		importance := defaultImportance
		if input.Importance != nil {
			importance = *input.Importance
		}

		item := db.ContextItem{
			ID:         uuid.NewString(),
			CreatedAt:  time.Now().Unix(),
			Source:     input.Source,
			ThreadID:   input.ThreadID,
			Role:       input.Role,
			Title:      input.Title,
			Content:    input.Content,
			Tags:       input.Tags,
			Importance: importance,
		}

		if err := store.InsertContext(ctx, item); err != nil {
			return nil, mcp.NewError(mcp.ErrInternal, err.Error())
		}

		return SaveContextResult{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
		}, nil
	}
}
