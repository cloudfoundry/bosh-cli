package platform_test

import (
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"

	boshdpresolv "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/infrastructure/devicepathresolver"
	fakedpresolv "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/infrastructure/devicepathresolver/fakes"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/platform"
	boshstats "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/platform/stats"
	fakestats "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/platform/stats/fakes"
	boshdirs "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/settings/directories"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system"
	fakesys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system/fakes"
)

var _ = Describe("DummyPlatform", describeDummyPlatform)

func describeDummyPlatform() {
	var (
		platform           Platform
		collector          boshstats.Collector
		fs                 boshsys.FileSystem
		cmdRunner          boshsys.CmdRunner
		dirProvider        boshdirs.Provider
		devicePathResolver boshdpresolv.DevicePathResolver
		logger             boshlog.Logger
	)

	BeforeEach(func() {
		collector = &fakestats.FakeCollector{}
		fs = fakesys.NewFakeFileSystem()
		cmdRunner = fakesys.NewFakeCmdRunner()
		dirProvider = boshdirs.NewProvider("/fake-dir")
		devicePathResolver = fakedpresolv.NewFakeDevicePathResolver()
		logger = boshlog.NewLogger(boshlog.LevelNone)
	})

	JustBeforeEach(func() {
		platform = NewDummyPlatform(
			collector,
			fs,
			cmdRunner,
			dirProvider,
			devicePathResolver,
			logger,
		)
	})

	Describe("GetDefaultNetwork", func() {
		It("returns the contents of dummy-defaults-network-settings.json since that's what the dummy cpi writes", func() {
			settingsFilePath := "/fake-dir/bosh/dummy-default-network-settings.json"
			fs.WriteFileString(settingsFilePath, `{"IP": "1.2.3.4"}`)

			network, err := platform.GetDefaultNetwork()
			Expect(err).NotTo(HaveOccurred())

			Expect(network.IP).To(Equal("1.2.3.4"))
		})
	})

	Describe("GetCertManager", func() {
		It("returs a dummy cert manager", func() {
			certManager := platform.GetCertManager()

			Expect(certManager.UpdateCertificates("")).Should(BeNil())
		})
	})
}
