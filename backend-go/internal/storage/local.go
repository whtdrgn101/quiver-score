package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Local is a filesystem-backed ObjectStore for docker-compose dev so the API
// can run without GCP credentials. Not for production.
type Local struct {
	root string
}

func NewLocal(root string) (*Local, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &Local{root: root}, nil
}

// metaSuffix is appended to the object filename to store its content type.
const metaSuffix = ".ct"

func (l *Local) path(key string) string {
	return filepath.Join(l.root, filepath.FromSlash(key))
}

func (l *Local) Put(ctx context.Context, key, contentType string, body io.Reader) error {
	full := l.path(key)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	f, err := os.Create(full)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, body); err != nil {
		return err
	}
	return os.WriteFile(full+metaSuffix, []byte(contentType), 0o644)
}

func (l *Local) Get(ctx context.Context, key string) (io.ReadCloser, ObjectMeta, error) {
	full := l.path(key)
	f, err := os.Open(full)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ObjectMeta{}, ErrNotFound
		}
		return nil, ObjectMeta{}, err
	}
	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, ObjectMeta{}, err
	}
	contentType := "application/octet-stream"
	if ct, err := os.ReadFile(full + metaSuffix); err == nil {
		contentType = string(ct)
	}
	return f, ObjectMeta{ContentType: contentType, Size: stat.Size()}, nil
}

func (l *Local) Delete(ctx context.Context, key string) error {
	full := l.path(key)
	if err := os.Remove(full); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Remove(full + metaSuffix); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (l *Local) DeletePrefix(ctx context.Context, prefix string) (int, error) {
	dir := l.path(prefix)
	count := 0
	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return filepath.SkipDir
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, metaSuffix) {
			return nil
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		_ = os.Remove(path + metaSuffix)
		count++
		return nil
	})
	if walkErr != nil && !errors.Is(walkErr, os.ErrNotExist) {
		return count, walkErr
	}
	_ = os.RemoveAll(dir)
	return count, nil
}
