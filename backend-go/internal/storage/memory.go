package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
)

type memoryObject struct {
	contentType string
	data        []byte
}

// Memory is an in-process ObjectStore for unit tests.
type Memory struct {
	mu      sync.RWMutex
	objects map[string]memoryObject
}

func NewMemory() *Memory {
	return &Memory{objects: make(map[string]memoryObject)}
}

func (m *Memory) Put(ctx context.Context, key, contentType string, body io.Reader) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[key] = memoryObject{contentType: contentType, data: data}
	return nil
}

func (m *Memory) Get(ctx context.Context, key string) (io.ReadCloser, ObjectMeta, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	obj, ok := m.objects[key]
	if !ok {
		return nil, ObjectMeta{}, ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(obj.data)), ObjectMeta{
		ContentType: obj.contentType,
		Size:        int64(len(obj.data)),
	}, nil
}

func (m *Memory) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, key)
	return nil
}

func (m *Memory) DeletePrefix(ctx context.Context, prefix string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for k := range m.objects {
		if strings.HasPrefix(k, prefix) {
			delete(m.objects, k)
			count++
		}
	}
	return count, nil
}
