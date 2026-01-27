package tools

import (
	"context"
	"encoding/json"
	"strings"

	"vcontext/internal/common"
	"vcontext/internal/db"
	"vcontext/internal/mcp"
)

const (
	defaultTopK        = 5
	maxTopK            = 50
	defaultMinImportance = 1
)

type SearchContextParams struct {
	Query         string  `json:"query"`
	TopK          *int    `json:"top_k"`
	ThreadID      *string `json:"thread_id"`
	MinImportance *int    `json:"min_importance"`
}

type SearchContextResult struct {
	Items []db.SearchResult `json:"items"`
}

func SearchContextHandler(store *db.DB) mcp.Handler {
	return func(ctx context.Context, params json.RawMessage) (any, *mcp.RPCError) {
		var input SearchContextParams
		if err := decodeParams(params, &input); err != nil {
			return nil, err
		}

		query := strings.TrimSpace(input.Query)
		if query == "" {
			return nil, mcp.NewError(mcp.ErrInvalidParams, "query is required")
		}

		topK := defaultTopK
		if input.TopK != nil {
			topK = *input.TopK
		}
		topK = common.ClampInt(topK, 1, maxTopK)

		minImportance := defaultMinImportance
		if input.MinImportance != nil {
			minImportance = *input.MinImportance
		}
		if minImportance < 1 {
			minImportance = defaultMinImportance
		}

		results, err := store.SearchContext(ctx, query, topK, input.ThreadID, minImportance)
		if err != nil {
			return nil, mcp.NewError(mcp.ErrInternal, err.Error())
		}

		return SearchContextResult{Items: results}, nil
	}
}
