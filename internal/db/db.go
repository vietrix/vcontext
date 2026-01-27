package db

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

var ErrNotFound = errors.New("context item not found")

type DB struct {
	conn   *sql.DB
	logger *log.Logger
}

func Open(path string, logger *log.Logger) (*DB, error) {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA temp_store=MEMORY;",
	}
	for _, pragma := range pragmas {
		if _, err := conn.Exec(pragma); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("apply pragma %q: %w", pragma, err)
		}
	}

	if _, err := conn.Exec(schemaSQL); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return &DB{conn: conn, logger: logger}, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) InsertContext(ctx context.Context, item ContextItem) error {
	tagsJSON, err := encodeTags(item.Tags)
	if err != nil {
		return err
	}

	_, err = d.conn.ExecContext(
		ctx,
		`INSERT INTO context_items (
			id, created_at, source, thread_id, role, title, content, tags, importance
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID,
		item.CreatedAt,
		item.Source,
		item.ThreadID,
		item.Role,
		item.Title,
		item.Content,
		tagsJSON,
		item.Importance,
	)
	if err != nil {
		return fmt.Errorf("insert context: %w", err)
	}

	return nil
}

func (d *DB) GetContext(ctx context.Context, id string) (*ContextItem, error) {
	var item ContextItem
	var source sql.NullString
	var threadID sql.NullString
	var role sql.NullString
	var title sql.NullString
	var tags sql.NullString

	row := d.conn.QueryRowContext(
		ctx,
		`SELECT id, created_at, source, thread_id, role, title, content, tags, importance
		 FROM context_items WHERE id = ?`,
		id,
	)

	if err := row.Scan(
		&item.ID,
		&item.CreatedAt,
		&source,
		&threadID,
		&role,
		&title,
		&item.Content,
		&tags,
		&item.Importance,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get context: %w", err)
	}

	item.Source = nullStringPtr(source)
	item.ThreadID = nullStringPtr(threadID)
	item.Role = nullStringPtr(role)
	item.Title = nullStringPtr(title)

	parsedTags, err := decodeTags(tags)
	if err != nil {
		return nil, err
	}
	item.Tags = parsedTags

	return &item, nil
}

func (d *DB) SearchContext(ctx context.Context, query string, topK int, threadID *string, minImportance int) ([]SearchResult, error) {
	if topK <= 0 {
		topK = 5
	}

	builder := strings.Builder{}
	builder.WriteString(`SELECT ci.id, ci.title, ci.source, ci.thread_id, ci.created_at, ci.importance,
		snippet(context_items_fts, 0, '', '', '...', 10) AS snippet
		FROM context_items_fts
		JOIN context_items ci ON ci.rowid = context_items_fts.rowid
		WHERE context_items_fts MATCH ? AND ci.importance >= ?`)

	args := []any{query, minImportance}
	if threadID != nil {
		builder.WriteString(" AND ci.thread_id = ?")
		args = append(args, *threadID)
	}

	builder.WriteString(" ORDER BY bm25(context_items_fts) LIMIT ?")
	args = append(args, topK)

	rows, err := d.conn.QueryContext(ctx, builder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("search context: %w", err)
	}
	defer rows.Close()

	results := make([]SearchResult, 0, topK)
	for rows.Next() {
		var result SearchResult
		var title sql.NullString
		var source sql.NullString
		var thread sql.NullString

		if err := rows.Scan(
			&result.ID,
			&title,
			&source,
			&thread,
			&result.CreatedAt,
			&result.Importance,
			&result.Snippet,
		); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}

		result.Title = nullStringPtr(title)
		result.Source = nullStringPtr(source)
		result.ThreadID = nullStringPtr(thread)
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search results: %w", err)
	}

	return results, nil
}

func encodeTags(tags *[]string) (*string, error) {
	if tags == nil {
		return nil, nil
	}
	data, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("encode tags: %w", err)
	}
	encoded := string(data)
	return &encoded, nil
}

func decodeTags(raw sql.NullString) (*[]string, error) {
	if !raw.Valid {
		return nil, nil
	}
	trimmed := strings.TrimSpace(raw.String)
	if trimmed == "" {
		return nil, nil
	}

	var tags []string
	if err := json.Unmarshal([]byte(trimmed), &tags); err != nil {
		return nil, fmt.Errorf("decode tags: %w", err)
	}

	return &tags, nil
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	converted := value.String
	return &converted
}
