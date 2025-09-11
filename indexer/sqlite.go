package indexer

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/3timeslazy/nix-search-tv/indexer/jsonstream"
	"github.com/3timeslazy/nix-search-tv/indexer/zstd"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	db *sql.DB
}

const createTable = `
CREATE TABLE IF NOT EXISTS packages (
    package TEXT PRIMARY KEY,
    content BLOB NOT NULL
) WITHOUT ROWID;

PRAGMA synchronous = OFF;
PRAGMA journal_mode = OFF;
`

// --- PRAGMA synchronous = OFF;
// --- PRAGMA journal_mode = OFF;

func NewSQLite(dir string) (*SQLite, error) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}

	path := filepath.Join(dir, "pkgs.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	_, err = db.Exec(createTable)
	if err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}

	return &SQLite{
		db: db,
	}, nil
}

func (indexer *SQLite) Index(pkgs io.Reader, keys io.Writer) error {
	// TODO: delete previous data

	tx, err := indexer.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	stmt, err := tx.Prepare("INSERT INTO packages (package, content) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("prepare sql: %w", err)
	}
	defer stmt.Close()

	buf := []byte{}

	err = jsonstream.ParsePackages(pkgs, func(name string, content []byte) error {
		buf = zstd.Compress(buf, content)

		_, err := stmt.Exec(name, buf)
		if err != nil {
			return fmt.Errorf("insert package: %w", err)
		}
		buf = buf[:0]

		keys.Write(append([]byte(name), '\n'))

		return nil
	})
	if err != nil {
		return fmt.Errorf("parse packages: %w", err)
	}

	return tx.Commit()
}

func (indexer *SQLite) Load(pkgName string) (json.RawMessage, error) {
	row := indexer.db.QueryRow("SELECT content FROM packages WHERE package = ?", pkgName)
	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("select package: %w", err)
	}

	comp := []byte{}
	err := row.Scan(&comp)
	if err != nil {
		return nil, fmt.Errorf("scan content: %w", err)
	}

	pkg, err := zstd.Decompress(nil, comp)
	if err != nil {
		return nil, fmt.Errorf("decompress: %w", err)
	}

	return pkg, nil
}

func (indexer *SQLite) Close() error {
	return indexer.db.Close()
}
