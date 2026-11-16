package releasedir

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

type FSGenerator struct {
	dirPath string
	fs      boshsys.FileSystem
}

type jobSpec struct {
	Name       string                 `yaml:"name"`
	Templates  map[string]string      `yaml:"templates"`
	Packages   []string               `yaml:"packages"`
	Properties map[string]interface{} `yaml:"properties"`
}

type packageSpec struct {
	Name         string   `yaml:"name"`
	Dependencies []string `yaml:"dependencies"`
	Files        []string `yaml:"files"`
}

func NewFSGenerator(dirPath string, fs boshsys.FileSystem) FSGenerator {
	return FSGenerator{dirPath: dirPath, fs: fs}
}

func (g FSGenerator) GenerateJob(options GenerateJobOptions) (err error) {
	name := options.Name
	jobDirPath := filepath.Join(g.dirPath, "jobs", name)

	if g.fs.FileExists(jobDirPath) {
		return bosherr.Errorf("Job '%s' at '%s' already exists", name, jobDirPath)
	}

	generatedTemplates := map[string]string{}

	if options.Monit {
		generatedTemplates["ctl.erb"] = "bin/ctl"
	} else {
		generatedTemplates["bpm.yml.erb"] = "config/bpm.yml"
		generatedTemplates["bin/"+name] = "bin/" + name
	}

	templates := map[string]string{}

	for base, dest := range generatedTemplates {
		templates[base] = dest
	}

	type customTemplate struct {
		base     string
		contents []byte
	}

	var customTemplates []customTemplate

	templateSrcs := sortedKeys(options.Templates)
	for _, src := range templateSrcs {
		base := filepath.Base(src)

		if _, found := generatedTemplates[base]; found {
			return bosherr.Errorf("Template '%s' conflicts with generated template '%s'", src, base)
		}

		if _, found := templates[base]; found {
			return bosherr.Errorf("Multiple templates resolve to '%s': '%s'", base, src)
		}

		contents, readErr := g.fs.ReadFile(src)
		if readErr != nil {
			return bosherr.WrapErrorf(readErr, "Reading template '%s'", src)
		}

		templates[base] = options.Templates[src]
		customTemplates = append(customTemplates, customTemplate{base: base, contents: contents})
	}

	err = g.fs.MkdirAll(jobDirPath, os.ModePerm)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating job '%s' dir", name)
	}

	// A failure partway through must not leave a partial job behind, otherwise
	// retrying reports that the job already exists.
	defer func() {
		if err != nil {
			g.fs.RemoveAll(jobDirPath) //nolint:errcheck
		}
	}()

	err = g.fs.MkdirAll(filepath.Join(jobDirPath, "templates"), os.ModePerm)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating job '%s' templates dir", name)
	}

	if options.Monit {
		err = g.fs.WriteFileString(filepath.Join(jobDirPath, "templates", "ctl.erb"), fmt.Sprintf(ctlTemplate, name, name))
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating job '%s' ctl template", name)
		}

		err = g.fs.WriteFileString(filepath.Join(jobDirPath, "monit"), fmt.Sprintf(monitTemplate, name, name, name, name))
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating job '%s' monit file", name)
		}
	} else {
		err = g.fs.WriteFileString(filepath.Join(jobDirPath, "templates", "bpm.yml.erb"), fmt.Sprintf(bpmConfigTemplate, name, name, name))
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating job '%s' BPM config template", name)
		}

		err = g.fs.MkdirAll(filepath.Join(jobDirPath, "templates", "bin"), os.ModePerm)
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating job '%s' templates bin dir", name)
		}

		err = g.fs.WriteFileString(filepath.Join(jobDirPath, "templates", "bin", name), bpmExecutableTemplate)
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating job '%s' executable template", name)
		}

		err = g.fs.WriteFileString(filepath.Join(jobDirPath, "monit"), fmt.Sprintf(bpmMonitTemplate, name, name, name, name, name))
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating job '%s' monit file", name)
		}
	}

	for _, template := range customTemplates {
		dstPath := filepath.Join(jobDirPath, "templates", template.base)
		err = g.fs.WriteFile(dstPath, template.contents)
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating job '%s' template '%s'", name, template.base)
		}
	}

	packages := options.Packages
	if packages == nil {
		packages = []string{}
	}

	spec := jobSpec{
		Name:       name,
		Templates:  templates,
		Packages:   packages,
		Properties: map[string]interface{}{},
	}

	err = g.writeYAML(filepath.Join(jobDirPath, "spec"), spec)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating job '%s' spec file", name)
	}

	return nil
}

func (g FSGenerator) GeneratePackage(options GeneratePackageOptions) (err error) {
	name := options.Name
	pkgDirPath := filepath.Join(g.dirPath, "packages", name)

	if g.fs.FileExists(pkgDirPath) {
		return bosherr.Errorf("Package '%s' at '%s' already exists", name, pkgDirPath)
	}

	err = g.fs.MkdirAll(pkgDirPath, os.ModePerm)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating package '%s' dir", name)
	}

	defer func() {
		if err != nil {
			g.fs.RemoveAll(pkgDirPath) //nolint:errcheck
		}
	}()

	files := options.Files
	if files == nil {
		files = []string{}
	}

	dependencies := options.Dependencies
	if dependencies == nil {
		dependencies = []string{}
	}

	spec := packageSpec{
		Name:         name,
		Dependencies: dependencies,
		Files:        files,
	}

	err = g.writeYAML(filepath.Join(pkgDirPath, "spec"), spec)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating package '%s' spec file", name)
	}

	err = g.fs.WriteFileString(filepath.Join(pkgDirPath, "packaging"), "set -e\n")
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating package '%s' packaging file", name)
	}

	return nil
}

func (g FSGenerator) writeYAML(path string, value interface{}) error {
	bytes, err := yaml.Marshal(value)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling YAML")
	}

	bytes = append([]byte("---\n"), bytes...)

	return g.fs.WriteFile(path, bytes)
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

const monitTemplate = `check process %s
  with pidfile /var/vcap/sys/run/%s/pid
  start program "/var/vcap/jobs/%s/bin/ctl start"
  stop program "/var/vcap/jobs/%s/bin/ctl stop"
  group vcap
`

const ctlTemplate = `#!/bin/bash

RUN_DIR=/var/vcap/sys/run/%s
LOG_DIR=/var/vcap/sys/log/%s
PIDFILE=${RUN_DIR}/pid

case $1 in

  start)
    mkdir -p $RUN_DIR $LOG_DIR
    chown -R vcap:vcap $RUN_DIR $LOG_DIR

    echo $$ > $PIDFILE

    # Replace this with the command that starts your process.
    exec sleep 600

    ;;

  stop)

    kill -9 $(cat $PIDFILE)

    rm -f $PIDFILE

    ;;

  *)
    echo "Usage: ctl {start|stop}" ;;

esac
`

const bpmConfigTemplate = `processes:
- name: %s
  executable: /var/vcap/jobs/%s/bin/%s
  args: []
  env: {}
`

const bpmExecutableTemplate = `#!/bin/bash

set -e

# Replace this with the command that starts your process.
exec sleep 600
`

const bpmMonitTemplate = `check process %s
  with pidfile /var/vcap/sys/run/bpm/%s/%s.pid
  start program "/var/vcap/jobs/bpm/bin/bpm start %s"
  stop program "/var/vcap/jobs/bpm/bin/bpm stop %s"
  group vcap
`
