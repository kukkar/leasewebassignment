package store

// RepositoryConfig contains settings for repository initialization.
// Providers can add additional fields for specific implementations.
type RepositoryConfig struct {
	UploadDir string
}

// NewRepository creates a repository from the given config.
// This function can be extended later to choose a different implementation.
func NewRepository(cfg RepositoryConfig) Repository {
	return NewInMemoryRepository(cfg.UploadDir)
}
