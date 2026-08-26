package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// Open 打开（或创建）SQLite 数据库并应用连接级 PRAGMA。
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000; PRAGMA foreign_keys=ON;"); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// Store 是 SQLite 仓储。
type Store struct {
	db *sql.DB
}

// New 用已打开连接构造仓储。
func New(db *sql.DB) *Store { return &Store{db: db} }

// DB 暴露底层连接供事务/批量使用。
func (s *Store) DB() *sql.DB { return s.db }
