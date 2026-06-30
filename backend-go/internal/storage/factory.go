package storage

import (
	"context"
	"fmt"

	"github.com/quiverscore/backend-go/internal/config"
)

// FromConfig constructs the ObjectStore selected by cfg.StorageBackend.
//
// Backends:
//   - "gcs":    Google Cloud Storage (requires cfg.GCSBucket)
//   - "local":  filesystem under cfg.LocalStoragePath (docker-compose dev)
//   - "memory": in-process map (tests only)
func FromConfig(ctx context.Context, cfg *config.Config) (ObjectStore, error) {
	switch cfg.StorageBackend {
	case "gcs":
		if cfg.GCSBucket == "" {
			return nil, fmt.Errorf("storage: GCS_BUCKET must be set when STORAGE_BACKEND=gcs")
		}
		return NewGCS(ctx, cfg.GCSBucket)
	case "local", "":
		return NewLocal(cfg.LocalStoragePath)
	case "memory":
		return NewMemory(), nil
	default:
		return nil, fmt.Errorf("storage: unknown STORAGE_BACKEND %q", cfg.StorageBackend)
	}
}
