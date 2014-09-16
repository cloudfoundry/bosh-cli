package release

type Release struct {
	Name    string
	Version string

	CommitHash         string
	UncommittedChanges bool

	Jobs          []Job
	Packages      []*Package
	ExtractedPath string
	TarballPath   string
}

func (r Release) FindJobByName(jobName string) (Job, bool) {
	for _, job := range r.Jobs {
		if job.Name == jobName {
			return job, true
		}
	}

	return Job{}, false
}

type Package struct {
	Name          string
	Fingerprint   string
	SHA1          string
	Dependencies  []*Package
	ExtractedPath string
}

func (p Package) String() string {
	return p.Name
}
