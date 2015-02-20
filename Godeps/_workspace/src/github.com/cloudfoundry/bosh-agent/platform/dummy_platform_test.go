package platform_test

import (
	"encoding/json"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	. "github.com/cloudfoundry/bosh-agent/platform"
	fakestats "github.com/cloudfoundry/bosh-agent/platform/stats/fakes"
	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
	boshdirs "github.com/cloudfoundry/bosh-agent/settings/directories"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
)

var _ = Describe("dummyPlatform", func() {
	var (
		collector   *fakestats.FakeCollector
		fs          *fakesys.FakeFileSystem
		cmdRunner   *fakesys.FakeCmdRunner
		dirProvider boshdirs.Provider
		platform    Platform
	)

	BeforeEach(func() {
		collector = &fakestats.FakeCollector{}
		fs = fakesys.NewFakeFileSystem()
		cmdRunner = fakesys.NewFakeCmdRunner()
		dirProvider = boshdirs.NewProvider("/fake-dir")
		logger := boshlog.NewLogger(boshlog.LevelNone)
		platform = NewDummyPlatform(collector, fs, cmdRunner, dirProvider, nil, logger)
	})

	Describe("GetDefaultNetwork", func() {
		Context("when default networks settings file is found", func() {
			expectedNetwork := boshsettings.Network{
				Default: []string{"fake-default"},
				DNS:     []string{"fake-dns-name"},
				IP:      "fake-ip-address",
				Netmask: "fake-netmask",
				Gateway: "fake-gateway",
				Mac:     "fake-mac-address",
			}

			BeforeEach(func() {
				settingsPath := filepath.Join(dirProvider.BoshDir(), "dummy-default-network-settings.json")

				expectedNetworkBytes, err := json.Marshal(expectedNetwork)
				Expect(err).ToNot(HaveOccurred())

				fs.WriteFile(settingsPath, expectedNetworkBytes)
			})

			It("returns network", func() {
				network, err := platform.GetDefaultNetwork()
				Expect(err).ToNot(HaveOccurred())
				Expect(network).To(Equal(expectedNetwork))
			})
		})

		Context("when default networks settings file is not found", func() {
			It("does not return error because dummy configuration allows no dynamic IP", func() {
				_, err := platform.GetDefaultNetwork()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
