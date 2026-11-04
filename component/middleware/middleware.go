package middleware

import (
	"context"
	"io"
	"sync"
)

type LsInfoRequest struct{ Path string }

type FileInfo struct {
	Path       string
	IsDir      bool
	Size       int64
	ModifiedAt string
}

type ReadRequest struct {
	FilePath string
	Offset   int
	Limit    int
}

type FileContent struct{ Content string }

type WriteRequest struct {
	FilePath string
	Content  string
}

type EditRequest struct {
	FilePath   string
	OldString  string
	NewString  string
	ReplaceAll bool
}

type GrepRequest struct {
	Pattern         string
	Path            string
	Glob            string
	FileType        string
	CaseInsensitive bool
	EnableMultiline bool
	AfterLines      int
	BeforeLines     int
}

type GrepMatch struct {
	Path    string
	Line    int
	Content string
}

type GlobInfoRequest struct {
	Path    string
	Pattern string
}

type ExecuteRequest struct {
	Command            string
	RunInBackendGround bool
}

type ExecuteResponse struct {
	Output   string
	ExitCode *int
}

type Backend interface {
	LsInfo(context.Context, *LsInfoRequest) ([]FileInfo, error)
	Read(context.Context, *ReadRequest) (*FileContent, error)
	Write(context.Context, *WriteRequest) error
	Edit(context.Context, *EditRequest) error
	GrepRaw(context.Context, *GrepRequest) ([]GrepMatch, error)
	GlobInfo(context.Context, *GlobInfoRequest) ([]FileInfo, error)
}

type StreamingShell interface {
	ExecuteStreaming(context.Context, *ExecuteRequest) (*StreamReader[*ExecuteResponse], error)
}

type streamItem[T any] struct {
	value T
	err   error
}

type StreamReader[T any] struct {
	items <-chan streamItem[T]
}

func (r *StreamReader[T]) Recv() (T, error) {
	item, ok := <-r.items
	if !ok {
		var zero T
		return zero, io.EOF
	}
	return item.value, item.err
}

type StreamWriter[T any] struct {
	items chan<- streamItem[T]
	once  *sync.Once
}

func (w *StreamWriter[T]) Send(value T, err error) error {
	w.items <- streamItem[T]{value: value, err: err}
	return nil
}

func (w *StreamWriter[T]) Close() {
	w.once.Do(func() { close(w.items) })
}

func Pipe[T any](buffer int) (*StreamReader[T], *StreamWriter[T]) {
	items := make(chan streamItem[T], buffer)
	return &StreamReader[T]{items: items}, &StreamWriter[T]{items: items, once: new(sync.Once)}
}
