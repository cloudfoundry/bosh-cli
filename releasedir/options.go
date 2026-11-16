package releasedir

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
