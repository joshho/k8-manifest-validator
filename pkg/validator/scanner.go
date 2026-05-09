package validator

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"sigs.k8s.io/yaml"
)

// ErrFileTooLarge is returned when a file exceeds the maximum buffer size
var ErrFileTooLarge = errors.New("file too large")

// Resource represents a discovered Kubernetes manifest resource
type Resource struct {
	Path  string
	Bytes []byte
}

// Signature returns a hash of the resource content for deduplication
func (r Resource) Signature() string {
	hash := sha256.Sum256(r.Bytes)
	return base64.URLEncoding.EncodeToString(hash[:])
}

// Scanner discovers Kubernetes manifest files in directories
type Scanner struct {
	Extensions     []string // file extensions to include (e.g., ".yaml", ".yml", ".json")
	IgnorePattern  []*regexp.Regexp
	MaxBufferSize  int64 // maximum buffer size (default 256MB)
	InitBufferSize int64 // initial buffer size (default 4MB)
}

// NewScanner creates a new Scanner with default settings
func NewScanner() *Scanner {
	return &Scanner{
		Extensions:     []string{".yaml", ".yml", ".json"},
		IgnorePattern:  make([]*regexp.Regexp, 0),
		MaxBufferSize:  256 * 1024 * 1024,
		InitBufferSize: 4 * 1024 * 1024,
	}
}

// Walk recursively discovers all manifest files in the given path
func (s *Scanner) Walk(path string) ([]Resource, error) {
	var resources []Resource

	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible files
		}

		// Check if path matches any ignore pattern
		relPath, err := filepath.Rel(path, walkPath)
		if err != nil {
			return nil
		}
		if s.isIgnored(relPath) {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Check file extension
		ext := strings.ToLower(filepath.Ext(walkPath))
		if !s.hasValidExtension(ext) {
			return nil
		}

		// Read file with auto-growing buffer
		data, err := s.readFile(walkPath)
		if err != nil {
			return nil // skip files we can't read
		}

		// Split multi-document YAML and add each as separate resource
		docs := SplitYAMLDocument(data)
		for _, doc := range docs {
			// Check if this is a List and expand it into individual items
			if listItems := expandList(doc); listItems != nil {
				for _, item := range listItems {
					resources = append(resources, Resource{
						Path:  walkPath,
						Bytes: item,
					})
				}
			} else {
				resources = append(resources, Resource{
					Path:  walkPath,
					Bytes: doc,
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resources, nil
}

// isIgnored checks if the path matches any ignore pattern
func (s *Scanner) isIgnored(path string) bool {
	for _, pattern := range s.IgnorePattern {
		if pattern.MatchString(path) {
			return true
		}
	}
	return false
}

// hasValidExtension checks if the file extension is valid
func (s *Scanner) hasValidExtension(ext string) bool {
	for _, validExt := range s.Extensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

// readFile reads a file with an auto-growing buffer
func (s *Scanner) readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := stat.Size()
	if fileSize > s.MaxBufferSize {
		return nil, &os.PathError{
			Op:   "read",
			Path: path,
			Err:  ErrFileTooLarge,
		}
	}

	// Use initial buffer size, grow if needed
	bufferSize := s.InitBufferSize
	if fileSize > bufferSize {
		bufferSize = fileSize
	}
	if bufferSize > s.MaxBufferSize {
		bufferSize = s.MaxBufferSize
	}

	// Read exactly fileSize bytes into a properly sized buffer
	// Use io.ReadAll on a LimitedReader to handle auto-growing
	limited := io.LimitReader(file, bufferSize)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// expandList expands a List kind into its individual item resources.
// This handles the Kubernetes List type which wraps other resources.
func expandList(data []byte) [][]byte {
	var doc map[string]interface{}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil
	}
	if doc["kind"] != "List" {
		return nil
	}

	items, ok := doc["items"].([]interface{})
	if !ok || len(items) == 0 {
		return nil
	}

	var result [][]byte
	for _, item := range items {
		itemBytes, err := yaml.Marshal(item)
		if err != nil {
			continue
		}
		result = append(result, itemBytes)
	}
	return result
}

// SplitYAMLDocument splits a YAML byte slice into individual documents
// Multi-document YAML uses --- as a separator
func SplitYAMLDocument(data []byte) [][]byte {
	// Normalize line endings
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))

	var documents [][]byte
	var currentDoc bytes.Buffer
	lines := bytes.Split(data, []byte("\n"))

	for i, line := range lines {
		// Check for document separator
		trimmed := bytes.TrimSpace(line)
		if bytes.Equal(trimmed, []byte("---")) {
			if currentDoc.Len() > 0 {
				docBytes := bytes.TrimSpace(currentDoc.Bytes())
				if len(docBytes) > 0 {
					// Make a copy since we reuse the buffer for subsequent documents
					docCopy := make([]byte, len(docBytes))
					copy(docCopy, docBytes)
					documents = append(documents, docCopy)
				}
				currentDoc.Reset()
			}
			continue
		}

		// Skip empty lines at the beginning if we're at the start
		if currentDoc.Len() == 0 && len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		currentDoc.Write(line)
		if i < len(lines)-1 {
			currentDoc.WriteByte('\n')
		}
	}

	// Don't forget the last document
	if currentDoc.Len() > 0 {
		docBytes := bytes.TrimSpace(currentDoc.Bytes())
		if len(docBytes) > 0 {
			// Make a copy since the buffer may be reused
			docCopy := make([]byte, len(docBytes))
			copy(docCopy, docBytes)
			documents = append(documents, docCopy)
		}
	}

	// If no documents found but we have content, return the whole content as one document
	if len(documents) == 0 && len(bytes.TrimSpace(data)) > 0 {
		documents = append(documents, bytes.TrimSpace(data))
	}

	return documents
}

// IsValidUTF8 checks if the data is valid UTF-8
func IsValidUTF8(data []byte) bool {
	return utf8.Valid(data)
}
