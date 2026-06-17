package state_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	biproperty "github.com/cloudfoundry/bosh-utils/property"

	. "github.com/cloudfoundry/bosh-cli/v7/deployment/instance/state"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	bireljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	bitemplate "github.com/cloudfoundry/bosh-cli/v7/templatescompiler"
)

func makeJob(name string, consumes []bireljob.LinkDef, provides []bireljob.LinkDef) bireljob.Job {
	j := bireljob.NewJob(NewResource(name, name+"-fp", nil))
	j.Consumes = consumes
	j.Provides = provides
	return *j
}

func makeManifestTmpl(name string, consumes map[string]bideplmanifest.ManifestConsumesEntry) bideplmanifest.ReleaseJobRef {
	return bideplmanifest.ReleaseJobRef{
		Name:     name,
		Release:  "fake-release",
		Consumes: consumes,
	}
}

var _ = Describe("ResolveLinks", func() {
	var (
		deploymentName     = "my-deployment"
		jobGroupName       = "bosh"
		defaultNetworkName = "default"
		allInstances       = []bitemplate.LinkInstanceSpec{
			{Name: "bosh", ID: "inst-0", Index: 0, Bootstrap: true, AZ: "z1", Address: "10.0.0.1"},
			{Name: "bosh", ID: "inst-1", Index: 1, Bootstrap: false, AZ: "z1", Address: "10.0.0.2"},
			{Name: "bosh", ID: "inst-2", Index: 2, Bootstrap: false, AZ: "z1", Address: "10.0.0.3"},
		}
		jobProps    = biproperty.Map{}
		globalProps = biproperty.Map{}
	)

	Describe("auto-resolution by type", func() {
		It("resolves a consumer to the single provider with matching type", func() {
			provider := makeJob("pxc-mysql", nil, []bireljob.LinkDef{
				{Name: "mysql", Type: "mysql"},
			})
			provider.Properties = map[string]bireljob.PropertyDefinition{
				"port": {Default: 3306},
			}
			provider.Provides[0] = bireljob.LinkDef{Name: "mysql", Type: "mysql", Properties: []string{"port"}}

			consumer := makeJob("galera-agent", []bireljob.LinkDef{
				{Name: "db", Type: "mysql"},
			}, nil)

			result, err := ResolveLinks(
				[]bireljob.Job{provider, consumer},
				[]bideplmanifest.ReleaseJobRef{
					makeManifestTmpl("pxc-mysql", nil),
					makeManifestTmpl("galera-agent", nil),
				},
				deploymentName, jobGroupName, defaultNetworkName, allInstances, jobProps, globalProps,
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveKey("galera-agent"))
			Expect(result["galera-agent"]).To(HaveKey("db"))

			spec := result["galera-agent"]["db"]
			Expect(spec.DeploymentName).To(Equal(deploymentName))
			Expect(spec.InstanceGroup).To(Equal(jobGroupName))
			Expect(spec.Instances).To(HaveLen(3))
			Expect(spec.Instances[0].Address).To(Equal("10.0.0.1"))
			Expect(spec.Properties["port"]).To(Equal(3306))
		})

		It("errors when the link is required and no provider is found", func() {
			consumer := makeJob("galera-agent", []bireljob.LinkDef{
				{Name: "db", Type: "mysql", Optional: false},
			}, nil)

			_, err := ResolveLinks(
				[]bireljob.Job{consumer},
				[]bideplmanifest.ReleaseJobRef{makeManifestTmpl("galera-agent", nil)},
				deploymentName, jobGroupName, defaultNetworkName, allInstances, jobProps, globalProps,
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("db"))
		})

		It("skips optional links when no provider is found", func() {
			consumer := makeJob("galera-agent", []bireljob.LinkDef{
				{Name: "db", Type: "mysql", Optional: true},
			}, nil)

			result, err := ResolveLinks(
				[]bireljob.Job{consumer},
				[]bideplmanifest.ReleaseJobRef{makeManifestTmpl("galera-agent", nil)},
				deploymentName, jobGroupName, defaultNetworkName, allInstances, jobProps, globalProps,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).NotTo(HaveKey("galera-agent"))
		})

		It("errors when multiple providers of the same type are found (ambiguous)", func() {
			provider1 := makeJob("pxc-mysql-a", nil, []bireljob.LinkDef{{Name: "mysql-a", Type: "mysql"}})
			provider2 := makeJob("pxc-mysql-b", nil, []bireljob.LinkDef{{Name: "mysql-b", Type: "mysql"}})
			consumer := makeJob("galera-agent", []bireljob.LinkDef{{Name: "db", Type: "mysql"}}, nil)

			_, err := ResolveLinks(
				[]bireljob.Job{provider1, provider2, consumer},
				[]bideplmanifest.ReleaseJobRef{
					makeManifestTmpl("pxc-mysql-a", nil),
					makeManifestTmpl("pxc-mysql-b", nil),
					makeManifestTmpl("galera-agent", nil),
				},
				deploymentName, jobGroupName, defaultNetworkName, allInstances, jobProps, globalProps,
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Ambiguous"))
		})
	})

	Describe("alias resolution via from:", func() {
		It("resolves a consumer to a provider by name using from:", func() {
			provider := makeJob("pxc-mysql", nil, []bireljob.LinkDef{
				{Name: "mysql", Type: "mysql"},
			})
			consumer := makeJob("galera-agent", []bireljob.LinkDef{
				{Name: "db", Type: "mysql"},
			}, nil)

			result, err := ResolveLinks(
				[]bireljob.Job{provider, consumer},
				[]bideplmanifest.ReleaseJobRef{
					makeManifestTmpl("pxc-mysql", nil),
					makeManifestTmpl("galera-agent", map[string]bideplmanifest.ManifestConsumesEntry{
						"db": {From: "mysql"},
					}),
				},
				deploymentName, jobGroupName, defaultNetworkName, allInstances, jobProps, globalProps,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result["galera-agent"]).To(HaveKey("db"))
		})

		It("errors when the named provider does not exist", func() {
			consumer := makeJob("galera-agent", []bireljob.LinkDef{
				{Name: "db", Type: "mysql"},
			}, nil)

			_, err := ResolveLinks(
				[]bireljob.Job{consumer},
				[]bideplmanifest.ReleaseJobRef{
					makeManifestTmpl("galera-agent", map[string]bideplmanifest.ManifestConsumesEntry{
						"db": {From: "no-such-provider"},
					}),
				},
				deploymentName, jobGroupName, defaultNetworkName, allInstances, jobProps, globalProps,
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no-such-provider"))
		})
	})

	Describe("blocked links", func() {
		It("omits the link when it is explicitly blocked in the manifest", func() {
			provider := makeJob("pxc-mysql", nil, []bireljob.LinkDef{{Name: "mysql", Type: "mysql"}})
			consumer := makeJob("galera-agent", []bireljob.LinkDef{
				{Name: "db", Type: "mysql"},
			}, nil)

			result, err := ResolveLinks(
				[]bireljob.Job{provider, consumer},
				[]bideplmanifest.ReleaseJobRef{
					makeManifestTmpl("pxc-mysql", nil),
					makeManifestTmpl("galera-agent", map[string]bideplmanifest.ManifestConsumesEntry{
						"db": {IsBlocked: true},
					}),
				},
				deploymentName, jobGroupName, defaultNetworkName, allInstances, jobProps, globalProps,
			)
			Expect(err).ToNot(HaveOccurred())
			// Either no entry for galera-agent, or the "db" key is absent.
			if links, ok := result["galera-agent"]; ok {
				Expect(links).NotTo(HaveKey("db"))
			}
		})
	})

	Describe("manual links", func() {
		It("builds a LinkSpec from manual instances/properties/address", func() {
			consumer := makeJob("director", []bireljob.LinkDef{
				{Name: "db", Type: "mysql"},
			}, nil)

			result, err := ResolveLinks(
				[]bireljob.Job{consumer},
				[]bideplmanifest.ReleaseJobRef{
					makeManifestTmpl("director", map[string]bideplmanifest.ManifestConsumesEntry{
						"db": {
							IsManual:   true,
							Instances:  []bideplmanifest.ManualLinkInstance{{Address: "1.2.3.4"}},
							Properties: map[string]interface{}{"port": 5432},
							Address:    "1.2.3.4",
						},
					}),
				},
				deploymentName, jobGroupName, defaultNetworkName, allInstances, jobProps, globalProps,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result["director"]).To(HaveKey("db"))

			spec := result["director"]["db"]
			Expect(spec.Address).To(Equal("1.2.3.4"))
			Expect(spec.Instances).To(HaveLen(1))
			Expect(spec.Instances[0].Address).To(Equal("1.2.3.4"))
			Expect(spec.Properties["port"]).To(Equal(5432))
		})
	})

	Describe("property extraction", func() {
		It("extracts link properties from the provider's manifest job properties first", func() {
			provider := bireljob.Job{}
			providerResource := NewResource("pxc-mysql", "pxc-mysql-fp", nil)
			provider = *bireljob.NewJob(providerResource)
			provider.Provides = []bireljob.LinkDef{
				{Name: "mysql", Type: "mysql", Properties: []string{"port"}},
			}
			provider.Properties = map[string]bireljob.PropertyDefinition{
				"port": {Default: 3306},
			}
			consumer := makeJob("galera-agent", []bireljob.LinkDef{{Name: "db", Type: "mysql"}}, nil)

			jobPortMap := biproperty.Map{"port": 13306}
			templates := []bideplmanifest.ReleaseJobRef{
				{Name: "pxc-mysql", Release: "pxc", Properties: &jobPortMap},
				makeManifestTmpl("galera-agent", nil),
			}

			result, err := ResolveLinks(
				[]bireljob.Job{provider, consumer},
				templates,
				deploymentName, jobGroupName, defaultNetworkName, allInstances, jobProps, globalProps,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result["galera-agent"]["db"].Properties["port"]).To(Equal(13306))
		})

		It("falls back to spec defaults when no manifest property is set", func() {
			provider := *bireljob.NewJob(NewResource("pxc-mysql", "fp", nil))
			provider.Provides = []bireljob.LinkDef{
				{Name: "mysql", Type: "mysql", Properties: []string{"port"}},
			}
			provider.Properties = map[string]bireljob.PropertyDefinition{
				"port": {Default: 3306},
			}
			consumer := makeJob("galera-agent", []bireljob.LinkDef{{Name: "db", Type: "mysql"}}, nil)

			result, err := ResolveLinks(
				[]bireljob.Job{provider, consumer},
				[]bideplmanifest.ReleaseJobRef{
					makeManifestTmpl("pxc-mysql", nil),
					makeManifestTmpl("galera-agent", nil),
				},
				deploymentName, jobGroupName, defaultNetworkName, allInstances, jobProps, globalProps,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result["galera-agent"]["db"].Properties["port"]).To(Equal(3306))
		})
	})

	Describe("AllInstanceSpecs", func() {
		It("builds instances from static IPs", func() {
			deploymentJob := bideplmanifest.Job{
				Name:      "bosh",
				Instances: 3,
				Networks: []bideplmanifest.JobNetwork{
					{Name: "default", StaticIPs: []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}},
				},
			}

			instances := AllInstanceSpecs(deploymentJob, "default")
			Expect(instances).To(HaveLen(3))
			Expect(instances[0].Bootstrap).To(BeTrue())
			Expect(instances[0].Address).To(Equal("10.0.0.1"))
			Expect(instances[1].Bootstrap).To(BeFalse())
			Expect(instances[1].Address).To(Equal("10.0.0.2"))
			Expect(instances[2].Address).To(Equal("10.0.0.3"))
		})

		It("reuses the first IP when there are fewer static IPs than instances", func() {
			deploymentJob := bideplmanifest.Job{
				Name:      "bosh",
				Instances: 3,
				Networks: []bideplmanifest.JobNetwork{
					{Name: "default", StaticIPs: []string{"10.0.0.1"}},
				},
			}

			instances := AllInstanceSpecs(deploymentJob, "default")
			Expect(instances).To(HaveLen(3))
			Expect(instances[0].Address).To(Equal("10.0.0.1"))
			Expect(instances[1].Address).To(Equal("10.0.0.1"))
			Expect(instances[2].Address).To(Equal("10.0.0.1"))
		})

		It("returns empty address when no static IPs are configured", func() {
			deploymentJob := bideplmanifest.Job{
				Name:      "bosh",
				Instances: 2,
				Networks:  []bideplmanifest.JobNetwork{{Name: "default"}},
			}

			instances := AllInstanceSpecs(deploymentJob, "default")
			Expect(instances).To(HaveLen(2))
			Expect(instances[0].Address).To(Equal(""))
		})
	})
})
