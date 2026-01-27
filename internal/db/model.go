package db

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
