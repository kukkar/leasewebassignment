package store

// RepositoryConfig contains settings for repository initialization.
type RepositoryConfig struct {
	// UploadDir is where uploaded CSV files are written for audit purposes.
	UploadDir string
}

// NewRepository creates the server repository. The only backend today is an
// in-memory store seeded from disk at startup (see service.LoadServerData);
// there is no requirement in the assignment for state to survive a process
// restart without that reload, so a persistent-storage backend was left out
// rather than shipped unwired and untested.
func NewRepository(cfg RepositoryConfig) Repository {
	return newMemoryRepository(cfg.UploadDir)
}
