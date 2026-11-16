package releasedir

// InitOptions describes how a release directory should be initialized.
type InitOptions struct {
	// Name is the release name. When empty it is derived from the directory
	// name with a trailing "-release" suffix removed.
	Name string

	Git    bool
	Sample bool
	Force  bool

	BlobstoreProvider       string
	BlobstoreOptions        map[string]interface{}
	BlobstorePrivateOptions map[string]interface{}
}

// GenerateJobOptions describes the job scaffold to generate.
type GenerateJobOptions struct {
	Name string

	// Monit generates a traditional monit-based job instead of the default
	// BPM-based job.
	Monit bool

	// Templates maps source file paths to their destination inside the job.
	Templates map[string]string

	Packages []string
}

// GeneratePackageOptions describes the package scaffold to generate.
type GeneratePackageOptions struct {
	Name string

	Files        []string
	Dependencies []string
}
