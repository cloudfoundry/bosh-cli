package releasedir

import (
	"path/filepath"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type FSGitRepo struct {
	// Will use CmdRunner's WorkingDir when exec-ing to avoid --git-dir weirdness
	dirPath string
	runner  boshsys.CmdRunner
	fs      boshsys.FileSystem
}

func NewFSGitRepo(dirPath string, runner boshsys.CmdRunner, fs boshsys.FileSystem) FSGitRepo {
	return FSGitRepo{dirPath: dirPath, runner: runner, fs: fs}
}

func (r FSGitRepo) Init() error {
	_, _, _, err := r.runner.RunCommand("git", "init", r.dirPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Initing git repo")
	}

	return r.WriteIgnoreFile()
}

func (r FSGitRepo) WriteIgnoreFile() error {
	ignoreTpl := `config/private.yml
blobs
.blobs
dev_releases
releases/*.tgz
releases/**/*.tgz
.dev_builds
.final_builds/jobs/**/*.tgz
.final_builds/packages/**/*.tgz
.DS_Store
.idea
*.swp
*~
*#
#*
`

	ignorePath := filepath.Join(r.dirPath, ".gitignore")

	if r.fs.FileExists(ignorePath) {
		existing, err := r.fs.ReadFileString(ignorePath)
		if err != nil {
			return bosherr.WrapError(err, "Reading .gitignore file")
		}

		merged := mergeIgnoreFile(existing, ignoreTpl)
		if merged == existing {
			return nil
		}

		err = r.fs.WriteFileString(ignorePath, merged)
		if err != nil {
			return bosherr.WrapError(err, "Updating .gitignore file")
		}

		return nil
	}

	err := r.fs.WriteFileString(ignorePath, ignoreTpl)
	if err != nil {
		return bosherr.WrapError(err, "Creating .gitignore file")
	}

	return nil
}

// mergeIgnoreFile appends template entries that are missing from the existing
// file while preserving custom entries.
func mergeIgnoreFile(existing, template string) string {
	known := map[string]bool{}

	for _, line := range strings.Split(existing, "\n") {
		known[strings.TrimSpace(line)] = true
	}

	var missing []string

	for _, line := range strings.Split(template, "\n") {
		entry := strings.TrimSpace(line)
		if len(entry) == 0 || known[entry] {
			continue
		}

		missing = append(missing, line)
		known[entry] = true
	}

	if len(missing) == 0 {
		return existing
	}

	lineEnding := "\n"
	if strings.Contains(existing, "\r\n") {
		lineEnding = "\r\n"
	}

	merged := existing
	if len(merged) > 0 && !strings.HasSuffix(merged, "\n") {
		merged += lineEnding
	}

	return merged + strings.Join(missing, lineEnding) + lineEnding
}

func (r FSGitRepo) LastCommitSHA() (string, error) {
	cmd := boshsys.Command{
		Name:       "git",
		Args:       []string{"rev-parse", "--short", "HEAD"},
		WorkingDir: r.dirPath,
	}
	stdout, stderr, _, err := r.runner.RunComplexCommand(cmd)
	if err != nil {
		if r.isNotGitRepo(stderr) {
			return "non-git", nil
		}

		if strings.Contains(stderr, "Needed a single revision") {
			return "empty", nil
		}

		return "", bosherr.WrapErrorf(err, "Checking last commit SHA")
	}

	return strings.TrimSpace(stdout), nil
}

func (r FSGitRepo) MustNotBeDirty(force bool) (bool, error) {
	cmd := boshsys.Command{
		Name:       "git",
		Args:       []string{"status", "--porcelain=1"},
		WorkingDir: r.dirPath,
	}
	stdout, stderr, _, err := r.runner.RunComplexCommand(cmd)
	if err != nil {
		if r.isNotGitRepo(stderr) {
			return false, nil
		}

		return false, bosherr.WrapErrorf(err, "Checking status")
	}

	// Strip out newline which is added if there are any changes
	stdout = strings.TrimSpace(stdout)

	if len(stdout) > 0 {
		if force {
			return true, nil
		}
		return true, bosherr.Errorf("Git repository has local modifications:\n\n%s", stdout)
	}

	return false, nil
}

func (r FSGitRepo) isNotGitRepo(stderr string) bool {
	if r.fs.FileExists(filepath.Join(r.dirPath, ".git")) {
		return false
	}

	cmd := boshsys.Command{
		Name:       "git",
		Args:       []string{"rev-parse", "--git-dir"},
		WorkingDir: r.dirPath,
	}
	_, _, _, err := r.runner.RunComplexCommand(cmd)
	return err != nil
}
