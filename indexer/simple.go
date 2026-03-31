package indexer

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/3timeslazy/nix-search-tv/indexer/jsonstream"
	"github.com/3timeslazy/nix-search-tv/indexer/zstd"
)

type Simple struct {
	indexPath string
	dataPath  string
}

func NewSimple(dir string) (*Simple, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir path: %w", err)
	}

	indexPath := filepath.Join(dir, "index.bin")
	dataPath := filepath.Join(dir, "data.bin")

	return &Simple{
		indexPath: indexPath,
		dataPath:  dataPath,
	}, nil
}

func (indexer *Simple) Index(pkgs io.Reader, indexedKeys io.Writer) error {
	pkgsIndex := make([]byte, 0)

	dataFile, err := os.OpenFile(indexer.dataPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0655)
	if err != nil {
		return fmt.Errorf("open data file: %w", err)
	}
	bufWr := bufio.NewWriter(dataFile)

	offset := 0
	nameb := make([]byte, 0, 255)
	compPkg := []byte{}

	err = jsonstream.ParsePackages(pkgs, func(name string, content []byte) error {
		nameb = []byte(name)

		compPkg = zstd.Compress(compPkg, content)
		l := len(compPkg)

		pkgsIndex = append(pkgsIndex, byte(len(name)))
		pkgsIndex = append(pkgsIndex, nameb...)
		pkgsIndex = binary.AppendVarint(pkgsIndex, int64(offset))
		pkgsIndex = binary.AppendVarint(pkgsIndex, int64(l))

		n, err := bufWr.Write(compPkg)
		if err != nil {
			return fmt.Errorf("write to data file: %w", err)
		}
		if n != l {
			return fmt.Errorf("bytes written not the same as content, %d != %d", n, l)
		}

		compPkg = compPkg[:0]
		offset += l

		indexedKeys.Write(append(nameb, '\n'))

		return nil
	})
	if err != nil {
		return fmt.Errorf("parse packages: %w", err)
	}

	if err = bufWr.Flush(); err != nil {
		return fmt.Errorf("flush data file: %w", err)
	}

	err = os.WriteFile(indexer.indexPath, pkgsIndex, 0655)
	if err != nil {
		return fmt.Errorf("write index file: %w", err)
	}

	return nil
}

func (indexer *Simple) Load(pkgName string) (json.RawMessage, error) {
	index, err := os.Open(indexer.indexPath)
	if err != nil {
		return nil, fmt.Errorf("read index file: %w", err)
	}
	defer index.Close()

	offset, length := int64(-1), int64(-1)

	rd := bufio.NewReader(index)
	for {
		slen, err := rd.ReadByte()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read string len: %w", err)
		}

		s := make([]byte, slen)
		n, err := io.ReadFull(rd, s)
		if err != nil {
			return nil, fmt.Errorf("read string: %w", err)
		}
		if n != len(s) {
			return nil, fmt.Errorf("expected string to be of length %d, but got %d", len(s), n)
		}

		off, err := binary.ReadVarint(rd)
		if err != nil {
			return nil, fmt.Errorf("read offset: %w", err)
		}
		l, err := binary.ReadVarint(rd)
		if err != nil {
			return nil, fmt.Errorf("read length: %w", err)
		}

		if string(s) == pkgName {
			offset = off
			length = l
			break
		}
	}
	if offset+length <= 0 {
		return nil, errors.New("not found")
	}

	dataFile, err := os.Open(indexer.dataPath)
	if err != nil {
		return nil, fmt.Errorf("open data file: %w", err)
	}
	defer dataFile.Close()

	compPkg := make([]byte, length)
	_, err = dataFile.ReadAt(compPkg, offset)
	if err != nil {
		return nil, fmt.Errorf("read package content: %w", err)
	}

	pkg, err := zstd.Decompress(nil, compPkg)
	if err != nil {
		return nil, fmt.Errorf("decompress content: %w", err)
	}

	return pkg, nil
}

func (indexer *Simple) Close() error {
	return nil
}
