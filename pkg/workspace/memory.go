package workspace

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"digital-labor/pkg/conf"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

const defaultAgentID = "default"

// DefaultAgentID is used by callers that have not yet been wired to an agent identity.
func DefaultAgentID() string {
	return defaultAgentID
}

// MemoryType defines the layers of memory.
type MemoryType string

const (
	MemTypeChunk    MemoryType = "chunk"
	MemTypeCompress MemoryType = "compress"
	MemTypeLongTerm MemoryType = "md"
)

// CompressionRecord describes one compressed summary segment and the original
// JSONL line range it summarizes. Line numbers are 1-based over LoadSession order.
type CompressionRecord struct {
	ID              string            `json:"id"`
	SourceStartLine int               `json:"source_start_line"`
	SourceEndLine   int               `json:"source_end_line"`
	Messages        []*schema.Message `json:"messages"`
	Created         time.Time         `json:"created"`
}

var (
	memoryMu sync.RWMutex
)

func getMemoryRoot() string {
	return filepath.Join(GetWorkspacePath(), "memory")
}

func getBaseName(agentID string, typ MemoryType) string {
	switch typ {
	case MemTypeChunk:
		return agentID
	case MemTypeCompress:
		return agentID + "-compress"
	case MemTypeLongTerm:
		return "memory"
	default:
		return ""
	}
}

func getPaths(agentID string, typ MemoryType) []string {
	root := getMemoryRoot()
	dir := filepath.Join(root, "sessions", agentID)
	base := getBaseName(agentID, typ)
	if base == "" {
		return nil
	}

	if typ == MemTypeLongTerm {
		path := filepath.Join(dir, base+".md")
		if _, err := os.Stat(path); err == nil {
			return []string{path}
		}
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, base) && strings.HasSuffix(name, ".jsonl") {
			// Ensure it's not another type's prefix
			if typ == MemTypeChunk && strings.Contains(name, "-compress") {
				continue
			}
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Strings(files)
	return files
}

// Save persists data for a given type. It appends for JSONL types and overwrites for md.
func Save(agentID string, typ MemoryType, data any) error {
	if err := validateID(agentID, "agent_id"); err != nil {
		return err
	}

	memoryMu.Lock()
	defer memoryMu.Unlock()

	root := getMemoryRoot()
	dir := filepath.Join(root, "sessions", agentID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	paths := getPaths(agentID, typ)
	var path string
	if len(paths) == 0 {
		base := getBaseName(agentID, typ)
		ext := ".jsonl"
		if typ == MemTypeLongTerm {
			ext = ".md"
		}
		path = filepath.Join(dir, base+ext)
	} else {
		path = paths[len(paths)-1]
		if typ != MemTypeLongTerm {
			if info, err := os.Stat(path); err == nil && info.Size() > conf.Conf.Memory.MaxMessagesSize {
				base := getBaseName(agentID, typ)
				path = filepath.Join(dir, fmt.Sprintf("%s.%d.jsonl", base, len(paths)))
			}
		}
	}

	switch typ {
	case MemTypeChunk:
		msg, ok := data.(*schema.Message)
		if !ok {
			return fmt.Errorf("invalid data type for %s: %T", typ, data)
		}
		return appendJSONL(path, msg)
	case MemTypeCompress:
		record, ok := data.(CompressionRecord)
		if !ok {
			recordPtr, ok := data.(*CompressionRecord)
			if !ok {
				return fmt.Errorf("invalid data type for %s: %T", typ, data)
			}
			record = *recordPtr
		}
		return appendJSONL(path, record)
	case MemTypeLongTerm:
		content, ok := data.(string)
		if !ok {
			return fmt.Errorf("invalid data type for %s: %T", typ, data)
		}
		return os.WriteFile(path, []byte(content), 0644)
	}

	return fmt.Errorf("unknown memory type: %s", typ)
}

// Load retrieves data for a given type.
func Load(agentID string, typ MemoryType) (any, error) {
	memoryMu.RLock()
	defer memoryMu.RUnlock()

	paths := getPaths(agentID, typ)
	if len(paths) == 0 {
		switch typ {
		case MemTypeChunk:
			return ([]*schema.Message)(nil), nil
		case MemTypeCompress:
			return ([]CompressionRecord)(nil), nil
		case MemTypeLongTerm:
			return "", nil
		}
		return nil, nil
	}

	switch typ {
	case MemTypeChunk:
		var messages []*schema.Message
		for _, p := range paths {
			msgs, err := loadJSONL[*schema.Message](p)
			if err != nil {
				return nil, err
			}
			messages = append(messages, msgs...)
		}
		return messages, nil
	case MemTypeCompress:
		var records []CompressionRecord
		for _, p := range paths {
			recs, err := loadJSONL[CompressionRecord](p)
			if err != nil {
				return nil, err
			}
			records = append(records, recs...)
		}
		return records, nil
	case MemTypeLongTerm:
		data, err := os.ReadFile(paths[0])
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	return nil, fmt.Errorf("unknown memory type: %s", typ)
}

// Delete removes all files for a given type.
func Delete(agentID string, typ MemoryType) error {
	memoryMu.Lock()
	defer memoryMu.Unlock()

	paths := getPaths(agentID, typ)
	for _, p := range paths {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// DeleteSession removes all files for an agent.
func DeleteSession(agentID string) error {
	memoryMu.Lock()
	defer memoryMu.Unlock()

	root := getMemoryRoot()
	dir := filepath.Join(root, "sessions", agentID)
	return os.RemoveAll(dir)
}

// Archive compresses files of a given type into a tar.gz archive and deletes the originals.
func Archive(agentID string, typ MemoryType) error {
	memoryMu.Lock()
	defer memoryMu.Unlock()

	paths := getPaths(agentID, typ)
	if len(paths) == 0 {
		return nil
	}

	root := getMemoryRoot()
	dir := filepath.Join(root, "sessions", agentID)
	archivePath := filepath.Join(dir, fmt.Sprintf("%s-%s.tar.gz", agentID, string(typ)))

	f, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, p := range paths {
		if err := addFileToTar(tw, p); err != nil {
			return err
		}
	}

	// Close writers before deleting files
	tw.Close()
	gw.Close()
	f.Close()

	for _, p := range paths {
		_ = os.Remove(p)
	}

	return nil
}

func addFileToTar(tw *tar.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = filepath.Base(path)

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}

// Helper functions

func appendJSONL(path string, data any) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}

func loadJSONL[T any](path string) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var results []T
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var item T
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}
		results = append(results, item)
	}
	return results, scanner.Err()
}

func validateID(id, name string) error {
	if id == "" {
		return fmt.Errorf("%s is empty", name)
	}
	if strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("%s contains path separator", name)
	}
	return nil
}

// Package-level helpers for LongTerm memory

func LoadAgentMemory(agentID string) string {
	res, err := Load(agentID, MemTypeLongTerm)
	if err != nil {
		return ""
	}
	return res.(string)
}

func SaveAgentMemory(agentID, content string) error {
	// Save for LongTerm is overwrite, so we need to load and append if we want to "save"
	old := LoadAgentMemory(agentID)
	return Save(agentID, MemTypeLongTerm, old+content)
}

func ReplaceAgentMemory(agentID, content string) error {
	return Save(agentID, MemTypeLongTerm, content)
}
