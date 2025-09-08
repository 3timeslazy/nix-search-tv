package indexer

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"

	"github.com/3timeslazy/nix-search-tv/indexer/x/jsonstream"
	"github.com/dgraph-io/badger/v4"
	"github.com/valyala/gozstd"
)

type Badger struct {
	badger *badger.DB
	ddict  *gozstd.DDict
	cdict  *gozstd.CDict
}

type BadgerConfig struct {
	Dir      string
	InMemory bool
}

func NewBadger(conf BadgerConfig) (*Badger, error) {
	opts := badger.
		DefaultOptions(conf.Dir).
		WithLoggingLevel(badger.ERROR).
		WithInMemory(conf.InMemory)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open badger: %w", err)
	}

	return &Badger{
		badger: db,

		cdict: cdict,
		ddict: ddict,
	}, nil
}

// Package defines fields set by the indexer during
// indexing
type Package struct {
	Name string `json:"_key"`
}

// Indexable represents the internal structure of the data
// that the indexer expects.
type Indexable struct {
	Packages map[string]json.RawMessage `json:"packages"`
}

func (indexer *Badger) Index(data io.Reader, indexedKeys io.Writer) error {
	// Delete previous index. If we do not do that
	// and just re-assign the keys below, the index
	// will be updated, however its size will increase drastically.
	// Then, to keep the index size small, we'll need to deal with
	// badger's garbade colletion. So, it's just easier to drop everything
	err := indexer.badger.DropAll()
	if err != nil {
		return fmt.Errorf("drop all: %w", err)
	}

	buf := []byte{}
	batch := indexer.badger.NewWriteBatch()

	err = jsonstream.ParsePackages(data, func(pkgName string, pkgContent []byte) error {
		nameb := []byte(pkgName)

		buf = gozstd.CompressDict(buf, pkgContent, indexer.cdict)
		err = batch.Set(nameb, bytes.Clone(buf))
		if err != nil {
			return fmt.Errorf("set package content: %w", err)
		}
		buf = buf[:0]

		indexedKeys.Write(append(nameb, []byte("\n")...))

		return nil
	})
	if err != nil {
		return fmt.Errorf("handle packages: %w", err)
	}

	return batch.Flush()
}

func (bdg *Badger) Load(pkgName string) (json.RawMessage, error) {
	pkg := []byte{}

	err := bdg.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(pkgName))
		if err != nil {
			return err
		}

		comp, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		pkg, err = gozstd.DecompressDict(pkg, comp, bdg.ddict)
		if err != nil {
			return fmt.Errorf("decompress content: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get package: %w", err)
	}

	return pkg, nil
}

func (bdg *Badger) Close() error {
	return bdg.badger.Close()
}
