package storage

import (
	"context"
	"errors"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// GCS is the production ObjectStore backed by Google Cloud Storage.
type GCS struct {
	client *storage.Client
	bucket string
}

func NewGCS(ctx context.Context, bucket string) (*GCS, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &GCS{client: client, bucket: bucket}, nil
}

func (g *GCS) Close() error {
	return g.client.Close()
}

func (g *GCS) obj(key string) *storage.ObjectHandle {
	return g.client.Bucket(g.bucket).Object(key)
}

func (g *GCS) Put(ctx context.Context, key, contentType string, body io.Reader) error {
	w := g.obj(key).NewWriter(ctx)
	w.ContentType = contentType
	if _, err := io.Copy(w, body); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func (g *GCS) Get(ctx context.Context, key string) (io.ReadCloser, ObjectMeta, error) {
	r, err := g.obj(key).NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, ObjectMeta{}, ErrNotFound
		}
		return nil, ObjectMeta{}, err
	}
	return r, ObjectMeta{ContentType: r.Attrs.ContentType, Size: r.Attrs.Size}, nil
}

func (g *GCS) Delete(ctx context.Context, key string) error {
	if err := g.obj(key).Delete(ctx); err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil
		}
		return err
	}
	return nil
}

// DeletePrefix lists every object under prefix and deletes them. Used for
// account deletion. Errors mid-iteration return what was deleted so far so
// callers can retry safely.
func (g *GCS) DeletePrefix(ctx context.Context, prefix string) (int, error) {
	bucket := g.client.Bucket(g.bucket)
	it := bucket.Objects(ctx, &storage.Query{Prefix: prefix})
	count := 0
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return count, err
		}
		if err := bucket.Object(attrs.Name).Delete(ctx); err != nil {
			if errors.Is(err, storage.ErrObjectNotExist) {
				continue
			}
			return count, err
		}
		count++
	}
	return count, nil
}
