// Package storage abstracts object storage so the rest of the API does not
// depend on Google Cloud Storage directly. Production wires the GCS backend;
// tests use the in-memory backend; local development can use the filesystem
// backend so docker compose works without GCP credentials.
package storage

import (
	"context"
	"errors"
	"io"
)

// ErrNotFound is returned by Get/Delete when an object does not exist.
var ErrNotFound = errors.New("storage: object not found")

// ObjectMeta describes an object returned from Get.
type ObjectMeta struct {
	ContentType string
	Size        int64
}

// ObjectStore is the contract every storage backend must satisfy. Keys are
// opaque strings — the application owns the layout (e.g. user-prefixed paths).
type ObjectStore interface {
	// Put writes an object. Existing objects with the same key are overwritten.
	Put(ctx context.Context, key, contentType string, body io.Reader) error

	// Get returns the object body and metadata. Caller must Close the reader.
	// Returns ErrNotFound if the key does not exist.
	Get(ctx context.Context, key string) (io.ReadCloser, ObjectMeta, error)

	// Delete removes a single object. Idempotent — missing keys do not error.
	Delete(ctx context.Context, key string) error

	// DeletePrefix removes every object under the given prefix and returns the
	// count of objects removed. Used for account deletion to wipe a user's
	// data in a single call.
	DeletePrefix(ctx context.Context, prefix string) (int, error)
}
