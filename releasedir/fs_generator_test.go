package releasedir_test

import (
	"errors"
	"os"
	"path/filepath"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/releasedir"
)

var _ = Describe("FSGenerator", func() {
	var (
		fs  *fakesys.FakeFileSystem
		gen FSGenerator
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		gen = NewFSGenerator("/dir", fs)
	})

	Describe("GenerateJob", func() {
		It("makes a BPM job by default", func() {
			err := gen.GenerateJob(GenerateJobOptions{Name: "job1"})
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "jobs", "job1", "spec"))).To(Equal(`---
name: job1
templates:
  bin/job1: bin/job1
  bpm.yml.erb: config/bpm.yml
packages: []
properties: {}
`))

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "jobs", "job1", "monit"))).To(Equal(`check process job1
  with pidfile /var/vcap/sys/run/bpm/job1/job1.pid
  start program "/var/vcap/jobs/bpm/bin/bpm start job1"
  stop program "/var/vcap/jobs/bpm/bin/bpm stop job1"
  group vcap
`))

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "jobs", "job1", "templates", "bpm.yml.erb"))).To(Equal(`processes:
- name: job1
  executable: /var/vcap/jobs/job1/bin/job1
  args: []
  env: {}
`))

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "jobs", "job1", "templates", "bin", "job1"))).To(Equal(`#!/bin/bash

set -e

# Replace this with the command that starts your process.
exec sleep 600
`))
		})

		It("makes a monit job when requested", func() {
			err := gen.GenerateJob(GenerateJobOptions{Name: "job1", Monit: true})
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "jobs", "job1", "spec"))).To(Equal(`---
name: job1
templates:
  ctl.erb: bin/ctl
packages: []
properties: {}
`))

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "jobs", "job1", "monit"))).To(Equal(`check process job1
  with pidfile /var/vcap/sys/run/job1/pid
  start program "/var/vcap/jobs/job1/bin/ctl start"
  stop program "/var/vcap/jobs/job1/bin/ctl stop"
  group vcap
`))

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "jobs", "job1", "templates", "ctl.erb"))).To(ContainSubstring("Usage: ctl {start|stop}"))
		})

		It("adds packages and custom templates", func() {
			err := fs.WriteFileString("/src/ctl.erb", "custom ctl")
			Expect(err).ToNot(HaveOccurred())

			err = gen.GenerateJob(GenerateJobOptions{
				Name:      "job1",
				Templates: map[string]string{"/src/ctl.erb": "bin/my-ctl"},
				Packages:  []string{"pkg1", "pkg2"},
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "jobs", "job1", "spec"))).To(Equal(`---
name: job1
templates:
  bin/job1: bin/job1
  bpm.yml.erb: config/bpm.yml
  ctl.erb: bin/my-ctl
packages:
- pkg1
- pkg2
properties: {}
`))

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "jobs", "job1", "templates", "ctl.erb"))).To(Equal("custom ctl"))
		})

		It("returns error if a custom template conflicts with a generated template", func() {
			err := fs.WriteFileString("/src/bpm.yml.erb", "custom")
			Expect(err).ToNot(HaveOccurred())

			err = gen.GenerateJob(GenerateJobOptions{
				Name:      "job1",
				Templates: map[string]string{"/src/bpm.yml.erb": "config/bpm.yml"},
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("conflicts with generated template"))
		})

		It("returns error if two custom templates share a name", func() {
			err := fs.WriteFileString("/src/a/notes.md", "one")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("/src/b/notes.md", "two")
			Expect(err).ToNot(HaveOccurred())

			err = gen.GenerateJob(GenerateJobOptions{
				Name: "job1",
				Templates: map[string]string{
					"/src/a/notes.md": "config/a.md",
					"/src/b/notes.md": "config/b.md",
				},
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Multiple templates resolve to 'notes.md'"))
		})

		It("returns error if job directory already exists", func() {
			err := fs.MkdirAll(filepath.Join("/", "dir", "jobs", "job1"), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			err = gen.GenerateJob(GenerateJobOptions{Name: "job1"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Job 'job1' at '" + filepath.Join("/", "dir", "jobs", "job1") + "' already exists"))
		})

		It("returns error if fails to create job directory", func() {
			fs.MkdirAllError = errors.New("fake-err")

			err := gen.GenerateJob(GenerateJobOptions{Name: "job1"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("does not leave a partial job when reading a template fails", func() {
			err := gen.GenerateJob(GenerateJobOptions{
				Name:      "job1",
				Templates: map[string]string{"/src/missing.erb": "config/missing"},
			})
			Expect(err).To(HaveOccurred())
			Expect(fs.FileExists(filepath.Join("/", "dir", "jobs", "job1"))).To(BeFalse())
		})

		It("does not leave a partial job when writing fails", func() {
			fs.WriteFileError = errors.New("fake-err")

			err := gen.GenerateJob(GenerateJobOptions{Name: "job1"})
			Expect(err).To(HaveOccurred())
			Expect(fs.FileExists(filepath.Join("/", "dir", "jobs", "job1"))).To(BeFalse())
		})
	})

	Describe("GeneratePackage", func() {
		It("makes package directory", func() {
			err := gen.GeneratePackage(GeneratePackageOptions{Name: "pkg1"})
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "packages", "pkg1", "spec"))).To(Equal(`---
name: pkg1
dependencies: []
files: []
`))

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "packages", "pkg1", "packaging"))).To(Equal("set -e\n"))
		})

		It("adds files and dependencies", func() {
			err := gen.GeneratePackage(GeneratePackageOptions{
				Name:         "pkg1",
				Files:        []string{"pkg1/**/*"},
				Dependencies: []string{"dep1"},
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString(filepath.Join("/", "dir", "packages", "pkg1", "spec"))).To(Equal(`---
name: pkg1
dependencies:
- dep1
files:
- pkg1/**/*
`))
		})

		It("returns error if package directory already exists", func() {
			err := fs.MkdirAll(filepath.Join("/", "dir", "packages", "pkg1"), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			err = gen.GeneratePackage(GeneratePackageOptions{Name: "pkg1"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Package 'pkg1' at '" + filepath.Join("/", "dir", "packages", "pkg1") + "' already exists"))
		})

		It("returns error if fails to create package directory", func() {
			fs.MkdirAllError = errors.New("fake-err")

			err := gen.GeneratePackage(GeneratePackageOptions{Name: "pkg1"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("does not leave a partial package when writing fails", func() {
			fs.WriteFileError = errors.New("fake-err")

			err := gen.GeneratePackage(GeneratePackageOptions{Name: "pkg1"})
			Expect(err).To(HaveOccurred())
			Expect(fs.FileExists(filepath.Join("/", "dir", "packages", "pkg1"))).To(BeFalse())
		})
	})
})
