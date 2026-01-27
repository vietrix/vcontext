package vcontext

import "encoding/json"

type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return e.Message
}

type ContextItem struct {
	ID         string    `json:"id"`
	CreatedAt  int64     `json:"created_at"`
	Source     *string   `json:"source,omitempty"`
	ThreadID   *string   `json:"thread_id,omitempty"`
	Role       *string   `json:"role,omitempty"`
	Title      *string   `json:"title,omitempty"`
	Content    string    `json:"content"`
	Tags       *[]string `json:"tags,omitempty"`
	Importance int       `json:"importance"`
}

type SearchResult struct {
	ID         string  `json:"id"`
	Title      *string `json:"title,omitempty"`
	Source     *string `json:"source,omitempty"`
	ThreadID   *string `json:"thread_id,omitempty"`
	CreatedAt  int64   `json:"created_at"`
	Importance int     `json:"importance"`
	Snippet    string  `json:"snippet"`
}

type SaveContextParams struct {
	Source     *string   `json:"source,omitempty"`
	ThreadID   *string   `json:"thread_id,omitempty"`
	Role       *string   `json:"role,omitempty"`
	Title      *string   `json:"title,omitempty"`
	Content    string    `json:"content"`
	Tags       *[]string `json:"tags,omitempty"`
	Importance *int      `json:"importance,omitempty"`
}

type SaveContextResult struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"created_at"`
}

type SearchContextParams struct {
	Query         string  `json:"query"`
	TopK          *int    `json:"top_k,omitempty"`
	ThreadID      *string `json:"thread_id,omitempty"`
	MinImportance *int    `json:"min_importance,omitempty"`
}

type SearchContextResult struct {
	Items []SearchResult `json:"items"`
}

type GetContextParams struct {
	ID string `json:"id"`
}
