CREATE TABLE IF NOT EXISTS context_items (
  id TEXT PRIMARY KEY,
  created_at INTEGER NOT NULL,
  source TEXT,
  thread_id TEXT,
  role TEXT,
  title TEXT,
  content TEXT NOT NULL,
  tags TEXT,
  importance INTEGER NOT NULL DEFAULT 3
);

CREATE VIRTUAL TABLE IF NOT EXISTS context_items_fts
USING fts5(
  content,
  title,
  tags,
  thread_id,
  content='',
  content_rowid='rowid'
);

CREATE TRIGGER IF NOT EXISTS context_items_ai AFTER INSERT ON context_items BEGIN
  INSERT INTO context_items_fts(rowid, content, title, tags, thread_id)
    VALUES (new.rowid, new.content, new.title, new.tags, new.thread_id);
END;

CREATE TRIGGER IF NOT EXISTS context_items_ad AFTER DELETE ON context_items BEGIN
  INSERT INTO context_items_fts(context_items_fts, rowid, content, title, tags, thread_id)
    VALUES('delete', old.rowid, old.content, old.title, old.tags, old.thread_id);
END;

CREATE TRIGGER IF NOT EXISTS context_items_au AFTER UPDATE ON context_items BEGIN
  INSERT INTO context_items_fts(context_items_fts, rowid, content, title, tags, thread_id)
    VALUES('delete', old.rowid, old.content, old.title, old.tags, old.thread_id);
  INSERT INTO context_items_fts(rowid, content, title, tags, thread_id)
    VALUES (new.rowid, new.content, new.title, new.tags, new.thread_id);
END;
