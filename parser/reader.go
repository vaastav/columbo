package parser

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
)

type Source int

type ProcessFn func(context.Context, string) error

const (
	LogFile Source = iota
	NamedPipe
)

type Reader struct {
	SrcType Source
	Path    string
}

func NewReader(ctx context.Context, path string, src Source) (*Reader, error) {
	return &Reader{SrcType: src, Path: path}, nil
}

func (r *Reader) ProcessLog(ctx context.Context, fn ProcessFn) error {
	if r.SrcType == LogFile {
		return r.ProcessFromLogFile(ctx, fn)
	}
	return errors.New(fmt.Sprintf("Unsupported source file type %d", int(r.SrcType)))
}

func (r *Reader) ProcessFromLogFile(ctx context.Context, fn ProcessFn) error {
	file, err := os.Open(r.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Handle empty log lines
		if line == "" {
			continue
		}
		err := fn(ctx, line)
		if err != nil {
			// Early exit if we find an error during processing
			return err
		}
	}
	return nil
}
