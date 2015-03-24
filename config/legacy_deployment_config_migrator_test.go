package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/config"

	//	"encoding/json"
	//	"errors"
	//
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
)

var _ = Describe("legacyDeploymentConfigMigrator", func() {
	var (
		migrator                       LegacyDeploymentConfigMigrator
		deploymentConfigService        DeploymentConfigService
		legacyDeploymentConfigFilePath string
		modernDeploymentConfigFilePath string
		fakeFs                         *fakesys.FakeFileSystem
		fakeUUIDGenerator              *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		fakeFs = fakesys.NewFakeFileSystem()
		fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
		legacyDeploymentConfigFilePath = "/path/to/legacy/bosh-deployment.yml"
		modernDeploymentConfigFilePath = "/path/to/legacy/deployment.json"
		logger := boshlog.NewLogger(boshlog.LevelNone)
		deploymentConfigService = NewFileSystemDeploymentConfigService(modernDeploymentConfigFilePath, fakeFs, fakeUUIDGenerator, logger)
		migrator = NewLegacyDeploymentConfigMigrator(legacyDeploymentConfigFilePath, deploymentConfigService, fakeFs, fakeUUIDGenerator, logger)
	})

	Describe("MigrateIfExists", func() {
		Context("when no legacy deploment config file exists", func() {
			It("does nothing", func() {
				migrated, err := migrator.MigrateIfExists()
				Expect(migrated).To(BeFalse())
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeFs.FileExists(modernDeploymentConfigFilePath)).To(BeFalse())
			})
		})

		Context("when legacy deploment config file exists (but is unparseable)", func() {
			BeforeEach(func() {
				fakeFs.WriteFileString(legacyDeploymentConfigFilePath, `xyz`)
			})

			It("does not delete the legacy deployment config file", func() {
				migrated, err := migrator.MigrateIfExists()
				Expect(migrated).To(BeFalse())
				Expect(err).To(HaveOccurred())

				Expect(fakeFs.FileExists(modernDeploymentConfigFilePath)).To(BeFalse())
				Expect(fakeFs.FileExists(legacyDeploymentConfigFilePath)).To(BeTrue())
			})
		})

		Context("when legacy deploment config file exists (and is empty)", func() {
			BeforeEach(func() {
				fakeFs.WriteFileString(legacyDeploymentConfigFilePath, `--- {}`)
			})

			It("deletes the legacy deployment config file", func() {
				migrated, err := migrator.MigrateIfExists()
				Expect(migrated).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeFs.FileExists(legacyDeploymentConfigFilePath)).To(BeFalse())
			})
		})

		Context("when legacy deploment config file exists (without vm, disk, or stemcell)", func() {
			BeforeEach(func() {
				fakeFs.WriteFileString(legacyDeploymentConfigFilePath, `---
instances:
- :id: 1
  :name: micro-robinson
  :uuid: bm-5480c6bb-3ba8-449a-a262-a2e75fbe5daf
  :stemcell_cid:
  :stemcell_sha1:
  :stemcell_name:
  :config_sha1: f9bdbc6cf6bf922f520ee9c45ed94a16a46dd972
  :vm_cid:
  :disk_cid:
disks: []
registry_instances:
- :id: 1
  :instance_id: i-a1624150
  :settings: '{}'
`)
			})

			It("deletes the legacy deployment config file", func() {
				migrated, err := migrator.MigrateIfExists()
				Expect(migrated).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeFs.FileExists(legacyDeploymentConfigFilePath)).To(BeFalse())
			})

			It("creates a new deployment config file without vm, disk, or stemcell", func() {
				migrated, err := migrator.MigrateIfExists()
				Expect(migrated).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				content, err := fakeFs.ReadFileString(modernDeploymentConfigFilePath)
				Expect(err).ToNot(HaveOccurred())

				Expect(content).To(MatchRegexp(`{
    "director_id": "fake-uuid-0",
    "installation_id": "",
    "current_vm_cid": "",
    "current_stemcell_id": "",
    "current_disk_id": "",
    "current_release_ids": null,
    "current_manifest_sha1": "",
    "disks": \[\],
    "stemcells": \[\],
    "releases": \[\]
}`))
			})
		})

		Context("when legacy deploment config file exists (with vm, disk & stemcell)", func() {
			BeforeEach(func() {
				fakeFs.WriteFileString(legacyDeploymentConfigFilePath, `---
instances:
- :id: 1
  :name: micro-robinson
  :uuid: bm-5480c6bb-3ba8-449a-a262-a2e75fbe5daf
  :stemcell_cid: ami-f2503e9a light
  :stemcell_sha1: 561b73dafc86454751db09855b0de7a89f0b4337
  :stemcell_name: light-bosh-stemcell-2807-aws-xen-ubuntu-trusty-go_agent
  :config_sha1: f9bdbc6cf6bf922f520ee9c45ed94a16a46dd972
  :vm_cid: i-a1624150
  :disk_cid: vol-565ed74d
disks: []
registry_instances:
- :id: 1
  :instance_id: i-a1624150
  :settings: '{}'
`)
			})

			It("deletes the legacy deployment config file", func() {
				migrated, err := migrator.MigrateIfExists()
				Expect(migrated).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeFs.FileExists(legacyDeploymentConfigFilePath)).To(BeFalse())
			})

			It("creates a new deployment config file with vm, disk & stemcell", func() {
				migrated, err := migrator.MigrateIfExists()
				Expect(migrated).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				content, err := fakeFs.ReadFileString(modernDeploymentConfigFilePath)
				Expect(err).ToNot(HaveOccurred())

				Expect(content).To(MatchRegexp(`{
    "director_id": "fake-uuid-0",
    "installation_id": "",
    "current_vm_cid": "i-a1624150",
    "current_stemcell_id": "",
    "current_disk_id": "fake-uuid-1",
    "current_release_ids": null,
    "current_manifest_sha1": "",
    "disks": \[
        {
            "id": "fake-uuid-1",
            "cid": "vol-565ed74d",
            "size": 0,
            "cloud_properties": {}
        }
    \],
    "stemcells": \[
        {
            "id": "fake-uuid-2",
            "name": "light-bosh-stemcell-2807-aws-xen-ubuntu-trusty-go_agent",
            "version": "",
            "cid": "ami-f2503e9a light"
        }
    \],
    "releases": \[\]
}`))
			})
		})

		Context("when legacy deploment config file exists (with vm only)", func() {
			BeforeEach(func() {
				fakeFs.WriteFileString(legacyDeploymentConfigFilePath, `---
instances:
- :id: 1
  :name: micro-robinson
  :uuid: bm-5480c6bb-3ba8-449a-a262-a2e75fbe5daf
  :stemcell_cid:
  :stemcell_sha1:
  :stemcell_name:
  :config_sha1: f9bdbc6cf6bf922f520ee9c45ed94a16a46dd972
  :vm_cid: i-a1624150
  :disk_cid:
disks: []
registry_instances:
- :id: 1
  :instance_id: i-a1624150
  :settings: '{}'
`)
			})

			It("deletes the legacy deployment config file", func() {
				migrated, err := migrator.MigrateIfExists()
				Expect(migrated).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeFs.FileExists(legacyDeploymentConfigFilePath)).To(BeFalse())
			})

			It("creates a new deployment config file with vm", func() {
				migrated, err := migrator.MigrateIfExists()
				Expect(migrated).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				content, err := fakeFs.ReadFileString(modernDeploymentConfigFilePath)
				Expect(err).ToNot(HaveOccurred())

				Expect(content).To(MatchRegexp(`{
    "director_id": "fake-uuid-0",
    "installation_id": "",
    "current_vm_cid": "i-a1624150",
    "current_stemcell_id": "",
    "current_disk_id": "",
    "current_release_ids": null,
    "current_manifest_sha1": "",
    "disks": \[\],
    "stemcells": \[\],
    "releases": \[\]
}`))
			})
		})

		Context("when legacy deploment config file exists (with disk only)", func() {
			BeforeEach(func() {
				fakeFs.WriteFileString(legacyDeploymentConfigFilePath, `---
instances:
- :id: 1
  :name: micro-robinson
  :uuid: bm-5480c6bb-3ba8-449a-a262-a2e75fbe5daf
  :stemcell_cid:
  :stemcell_sha1:
  :stemcell_name:
  :config_sha1: f9bdbc6cf6bf922f520ee9c45ed94a16a46dd972
  :vm_cid:
  :disk_cid: vol-565ed74d
disks: []
registry_instances:
- :id: 1
  :instance_id: i-a1624150
  :settings: '{}'
`)
			})

			It("deletes the legacy deployment config file", func() {
				migrated, err := migrator.MigrateIfExists()
				Expect(migrated).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeFs.FileExists(legacyDeploymentConfigFilePath)).To(BeFalse())
			})

			It("creates a new deployment config file with disk only", func() {
				migrated, err := migrator.MigrateIfExists()
				Expect(migrated).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				content, err := fakeFs.ReadFileString(modernDeploymentConfigFilePath)
				Expect(err).ToNot(HaveOccurred())

				Expect(content).To(MatchRegexp(`{
    "director_id": "fake-uuid-0",
    "installation_id": "",
    "current_vm_cid": "",
    "current_stemcell_id": "",
    "current_disk_id": "fake-uuid-1",
    "current_release_ids": null,
    "current_manifest_sha1": "",
    "disks": \[
        {
            "id": "fake-uuid-1",
            "cid": "vol-565ed74d",
            "size": 0,
            "cloud_properties": {}
        }
    \],
    "stemcells": \[\],
    "releases": \[\]
}`))
			})
		})

		Context("when legacy deploment config file exists (with stemcell only)", func() {
			BeforeEach(func() {
				fakeFs.WriteFileString(legacyDeploymentConfigFilePath, `---
instances:
- :id: 1
  :name: micro-robinson
  :uuid: bm-5480c6bb-3ba8-449a-a262-a2e75fbe5daf
  :stemcell_cid: ami-f2503e9a light
  :stemcell_sha1: 561b73dafc86454751db09855b0de7a89f0b4337
  :stemcell_name: light-bosh-stemcell-2807-aws-xen-ubuntu-trusty-go_agent
  :config_sha1: f9bdbc6cf6bf922f520ee9c45ed94a16a46dd972
  :vm_cid:
  :disk_cid:
disks: []
registry_instances:
- :id: 1
  :instance_id: i-a1624150
  :settings: '{}'
`)
			})

			It("deletes the legacy deployment config file", func() {
				migrated, err := migrator.MigrateIfExists()
				Expect(migrated).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeFs.FileExists(legacyDeploymentConfigFilePath)).To(BeFalse())
			})

			It("creates a new deployment config file with stemcell only (marked as unused)", func() {
				migrated, err := migrator.MigrateIfExists()
				Expect(migrated).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				content, err := fakeFs.ReadFileString(modernDeploymentConfigFilePath)
				Expect(err).ToNot(HaveOccurred())

				Expect(content).To(MatchRegexp(`{
    "director_id": "fake-uuid-0",
    "installation_id": "",
    "current_vm_cid": "",
    "current_stemcell_id": "",
    "current_disk_id": "",
    "current_release_ids": null,
    "current_manifest_sha1": "",
    "disks": \[\],
    "stemcells": \[
        {
            "id": "fake-uuid-1",
            "name": "light-bosh-stemcell-2807-aws-xen-ubuntu-trusty-go_agent",
            "version": "",
            "cid": "ami-f2503e9a light"
        }
    \],
    "releases": \[\]
}`))
			})
		})
	})
})
