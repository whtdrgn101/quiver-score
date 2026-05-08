package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"testing"
)

// runConformanceSuite exercises any ObjectStore implementation against the
// behavior the rest of the API relies on.
func runConformanceSuite(t *testing.T, store ObjectStore) {
	t.Helper()
	ctx := context.Background()

	t.Run("PutGetRoundtrip", func(t *testing.T) {
		body := []byte("hello world")
		if err := store.Put(ctx, "users/u1/a/1.bin", "text/plain", bytes.NewReader(body)); err != nil {
			t.Fatalf("put: %v", err)
		}
		r, meta, err := store.Get(ctx, "users/u1/a/1.bin")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		defer r.Close()
		got, err := io.ReadAll(r)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if !bytes.Equal(got, body) {
			t.Errorf("got %q want %q", got, body)
		}
		if meta.ContentType != "text/plain" {
			t.Errorf("content type = %q", meta.ContentType)
		}
		if meta.Size != int64(len(body)) {
			t.Errorf("size = %d, want %d", meta.Size, len(body))
		}
	})

	t.Run("GetMissingReturnsNotFound", func(t *testing.T) {
		_, _, err := store.Get(ctx, "users/u1/missing.bin")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("DeleteIsIdempotent", func(t *testing.T) {
		if err := store.Delete(ctx, "users/u1/never-existed.bin"); err != nil {
			t.Errorf("delete missing: %v", err)
		}
		_ = store.Put(ctx, "users/u1/temp.bin", "text/plain", bytes.NewReader([]byte("x")))
		if err := store.Delete(ctx, "users/u1/temp.bin"); err != nil {
			t.Errorf("delete: %v", err)
		}
		if _, _, err := store.Get(ctx, "users/u1/temp.bin"); !errors.Is(err, ErrNotFound) {
			t.Errorf("after delete got %v, want ErrNotFound", err)
		}
	})

	t.Run("DeletePrefixOnlyMatchesUser", func(t *testing.T) {
		_ = store.Put(ctx, "users/userA/x/1.bin", "text/plain", bytes.NewReader([]byte("a1")))
		_ = store.Put(ctx, "users/userA/y/2.bin", "text/plain", bytes.NewReader([]byte("a2")))
		_ = store.Put(ctx, "users/userB/z/3.bin", "text/plain", bytes.NewReader([]byte("b1")))

		count, err := store.DeletePrefix(ctx, "users/userA/")
		if err != nil {
			t.Fatalf("delete prefix: %v", err)
		}
		if count != 2 {
			t.Errorf("count = %d, want 2", count)
		}
		if _, _, err := store.Get(ctx, "users/userA/x/1.bin"); !errors.Is(err, ErrNotFound) {
			t.Errorf("userA/x still present: %v", err)
		}
		if _, _, err := store.Get(ctx, "users/userB/z/3.bin"); err != nil {
			t.Errorf("userB unexpectedly affected: %v", err)
		}
	})
}

func TestMemory(t *testing.T) {
	runConformanceSuite(t, NewMemory())
}

func TestLocal(t *testing.T) {
	root := filepath.Join(t.TempDir(), "storage")
	store, err := NewLocal(root)
	if err != nil {
		t.Fatalf("new local: %v", err)
	}
	runConformanceSuite(t, store)
}
