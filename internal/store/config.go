package store

// RepositoryConfig contains settings for repository initialization.
// Providers can add additional fields for specific implementations.
type RepositoryConfig struct {
	UploadDir string
	// Kind selects repository implementation: "memory" (default) or "file".
	Kind string
}

// NewRepository creates a repository from the given config.
// This function can be extended later to choose a different implementation.
func NewRepository(cfg RepositoryConfig) Repository {
	switch cfg.Kind {
	case "", "memory":
		return NewInMemoryRepository(cfg.UploadDir)
	case "file":
		return NewFileRepository(cfg.UploadDir)
	default:
		return NewInMemoryRepository(cfg.UploadDir)
	}
}

// NewFileRepository currently returns an in-memory repository with a persisted upload directory.
// This is a placeholder for a full file-backed repository implementation.
func NewFileRepository(uploadDir string) Repository {
	return NewInMemoryRepository(uploadDir)
}
