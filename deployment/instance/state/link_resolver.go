package state

import (
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	biproperty "github.com/cloudfoundry/bosh-utils/property"

	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	bireljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	bitemplate "github.com/cloudfoundry/bosh-cli/v7/templatescompiler"
)

// ResolveLinks resolves all BOSH links for an instance group.
// It returns a nested map: job-template-name → link-name → LinkSpec.
// This map is stored in InstanceSpec.Links and passed as JSON to erb_renderer.rb.
//
// Resolution order for each consumer link:
//  1. Blocked (manifest: consumes: {db: ~}) → link absent from output map
//  2. Manual  (manifest has instances/properties/address) → built from manifest data
//  3. Alias   (manifest has from: <provider-name>) → find provider by name
//  4. Auto    → scan all jobs in instance group for a single matching type
//  5. Optional and no provider found → absent from output map (no error)
//  6. Required and no provider found → error
func ResolveLinks(
	releaseJobs []bireljob.Job,
	manifestTemplates []bideplmanifest.ReleaseJobRef,
	deploymentName string,
	jobGroupName string,
	defaultNetworkName string,
	allInstances []bitemplate.LinkInstanceSpec,
	jobProperties biproperty.Map,
	globalProperties biproperty.Map,
) (map[string]map[string]bitemplate.LinkSpec, error) {

	// Index manifest templates by job name for quick override lookup.
	manifestByName := make(map[string]bideplmanifest.ReleaseJobRef, len(manifestTemplates))
	for _, tmpl := range manifestTemplates {
		manifestByName[tmpl.Name] = tmpl
	}

	// Build provider indexes: by type (for auto-resolution) and by name (for aliases).
	type providerEntry struct {
		job      bireljob.Job
		provides bireljob.LinkDef
	}
	providersByType := make(map[string][]providerEntry)
	providersByName := make(map[string]providerEntry)

	for _, j := range releaseJobs {
		for _, p := range j.Provides {
			entry := providerEntry{job: j, provides: p}
			providersByType[p.Type] = append(providersByType[p.Type], entry)
			// providers are indexed by both the provides.name and the job name (as a fallback)
			providersByName[p.Name] = entry
		}
	}

	result := make(map[string]map[string]bitemplate.LinkSpec)

	for _, consumerJob := range releaseJobs {
		if len(consumerJob.Consumes) == 0 {
			continue
		}

		manifestTmpl := manifestByName[consumerJob.Name()]
		jobLinks := make(map[string]bitemplate.LinkSpec)

		for _, c := range consumerJob.Consumes {
			// Look up manifest-level override for this link.
			var override bideplmanifest.ManifestConsumesEntry
			var hasOverride bool
			if manifestTmpl.Consumes != nil {
				override, hasOverride = manifestTmpl.Consumes[c.Name]
			}

			// Step 1: blocked link.
			if hasOverride && override.IsBlocked {
				continue
			}

			// Step 2: manual link.
			if hasOverride && override.IsManual {
				spec := buildManualLinkSpec(deploymentName, jobGroupName, defaultNetworkName, override)
				jobLinks[c.Name] = spec
				continue
			}

			// Steps 3-4: find a provider job.
			var provider *providerEntry

			if hasOverride && override.From != "" {
				// Step 3: alias — look up by provider name.
				if p, ok := providersByName[override.From]; ok {
					p := p
					provider = &p
				} else {
					return nil, bosherr.Errorf(
						"Link '%s' in job '%s' uses from: '%s' but no provider with that name was found in the instance group",
						c.Name, consumerJob.Name(), override.From,
					)
				}
			} else {
				// Step 4: auto-resolve by type.
				candidates := providersByType[c.Type]
				switch len(candidates) {
				case 1:
					p := candidates[0]
					provider = &p
				case 0:
					// handled below
				default:
					return nil, bosherr.Errorf(
						"Ambiguous link '%s' (type '%s') in job '%s': %d providers found; "+
							"use 'consumes: {%s: {from: <provider-name>}}' in the manifest to disambiguate",
						c.Name, c.Type, consumerJob.Name(), len(candidates), c.Name,
					)
				}
			}

			// Step 5: no provider found — optional skips, required errors.
			if provider == nil {
				if c.Optional {
					continue
				}
				return nil, bosherr.Errorf(
					"Link '%s' (type '%s') in job '%s' is required but no provider was found in the instance group",
					c.Name, c.Type, consumerJob.Name(),
				)
			}

			// Extract the whitelisted properties from the provider job.
			linkProperties, err := extractLinkProperties(
				provider.job,
				manifestByName[provider.job.Name()],
				provider.provides,
				jobProperties,
				globalProperties,
			)
			if err != nil {
				return nil, bosherr.WrapErrorf(err,
					"Extracting link properties for '%s' provided by '%s'", c.Name, provider.job.Name(),
				)
			}

			spec := bitemplate.LinkSpec{
				DeploymentName:       deploymentName,
				Domain:               "bosh",
				InstanceGroup:        jobGroupName,
				DefaultNetwork:       defaultNetworkName,
				GroupName:            fmt.Sprintf("%s.%s.%s.bosh", provider.provides.Name, jobGroupName, deploymentName),
				Instances:            allInstances,
				Properties:           linkProperties,
				UseLinkDNSNames:      false,
				UseShortDNSAddresses: false,
			}
			jobLinks[c.Name] = spec
		}

		if len(jobLinks) > 0 {
			result[consumerJob.Name()] = jobLinks
		}
	}

	return result, nil
}

// buildManualLinkSpec constructs a LinkSpec from a manual link override in the manifest.
// A top-level Address field signals erb_renderer.rb to use ManualLinkDnsEncoder.
func buildManualLinkSpec(
	deploymentName string,
	jobGroupName string,
	defaultNetworkName string,
	override bideplmanifest.ManifestConsumesEntry,
) bitemplate.LinkSpec {
	instances := make([]bitemplate.LinkInstanceSpec, 0, len(override.Instances))
	for i, inst := range override.Instances {
		instances = append(instances, bitemplate.LinkInstanceSpec{
			Name:      jobGroupName,
			ID:        fmt.Sprintf("manual-%d", i),
			Index:     i,
			Bootstrap: i == 0,
			AZ:        "",
			Address:   inst.Address,
		})
	}

	props := make(map[string]interface{}, len(override.Properties))
	for k, v := range override.Properties {
		props[k] = v
	}

	return bitemplate.LinkSpec{
		DeploymentName:       deploymentName,
		Domain:               "bosh",
		InstanceGroup:        jobGroupName,
		DefaultNetwork:       defaultNetworkName,
		GroupName:            jobGroupName,
		Instances:            instances,
		Properties:           props,
		Address:              override.Address,
		UseLinkDNSNames:      false,
		UseShortDNSAddresses: false,
	}
}

// extractLinkProperties collects the properties that a provider exposes through a link.
// It mirrors LinksParser::LinkHelpers#process_link_properties from the BOSH director.
//
// Property values are resolved in priority order:
//  1. Provider job's manifest-level template properties (job-template overrides)
//  2. Instance group (deployment job) properties
//  3. Global deployment properties
//  4. Release job spec defaults
//
// Only properties listed in provides.Properties are included in the output.
// Properties are stored as nested maps (matching the BOSH director's behaviour) so that
// the ERB renderer's dotted-path lookup_property function can traverse them.
func extractLinkProperties(
	providerJob bireljob.Job,
	providerManifestTmpl bideplmanifest.ReleaseJobRef,
	provides bireljob.LinkDef,
	jobProperties biproperty.Map,
	globalProperties biproperty.Map,
) (map[string]interface{}, error) {
	if len(provides.Properties) == 0 {
		return map[string]interface{}{}, nil
	}

	// Build spec defaults map, expanding flat dotted property names (e.g.
	// "endpoint_tls.enabled") into nested maps so that lookupNestedProperty can
	// traverse them.  Job spec files use flat dotted keys; manifests use nested YAML.
	specDefaults := make(biproperty.Map)
	for name, def := range providerJob.Properties {
		if def.Default != nil {
			setNestedPropertyBiproperty(specDefaults, name, def.Default)
		}
	}

	result := make(map[string]interface{})
	for _, propName := range provides.Properties {
		val, found := resolveProperty(propName, providerManifestTmpl.Properties, jobProperties, globalProperties, specDefaults)
		if !found {
			// Property not set anywhere; omit it (director also omits missing properties).
			continue
		}
		// Store as nested map so the ERB renderer's dotted-path lookup works:
		// "a.b.c" => val becomes {"a": {"b": {"c": val}}}
		setNestedProperty(result, propName, val)
	}
	return result, nil
}

// setNestedProperty stores val at the dotted path within m, creating intermediate
// maps as needed.  For example, setNestedProperty(m, "a.b.c", 1) produces
// m["a"]["b"]["c"] = 1.
func setNestedProperty(m map[string]interface{}, dotPath string, val interface{}) {
	parts := strings.SplitN(dotPath, ".", 2)
	if len(parts) == 1 {
		m[parts[0]] = val
		return
	}
	nested, ok := m[parts[0]].(map[string]interface{})
	if !ok {
		nested = make(map[string]interface{})
		m[parts[0]] = nested
	}
	setNestedProperty(nested, parts[1], val)
}

// setNestedPropertyBiproperty is like setNestedProperty but works with biproperty.Map.
// It is used to expand flat dotted property names from job spec defaults into the
// nested map structure expected by lookupNestedProperty.
func setNestedPropertyBiproperty(m biproperty.Map, dotPath string, val interface{}) {
	parts := strings.SplitN(dotPath, ".", 2)
	if len(parts) == 1 {
		m[parts[0]] = val
		return
	}
	existing := m[parts[0]]
	nested, ok := existing.(biproperty.Map)
	if !ok {
		nested = make(biproperty.Map)
		m[parts[0]] = nested
	}
	setNestedPropertyBiproperty(nested, parts[1], val)
}

// resolveProperty resolves a single (possibly dot-nested) property name through the chain:
//   templateProperties → jobProperties → globalProperties → specDefaults
func resolveProperty(
	name string,
	templateProperties *biproperty.Map,
	jobProperties biproperty.Map,
	globalProperties biproperty.Map,
	specDefaults biproperty.Map,
) (interface{}, bool) {
	if templateProperties != nil {
		if val, found := lookupNestedProperty(*templateProperties, name); found {
			return val, true
		}
	}
	if val, found := lookupNestedProperty(jobProperties, name); found {
		return val, true
	}
	if val, found := lookupNestedProperty(globalProperties, name); found {
		return val, true
	}
	if val, found := lookupNestedProperty(specDefaults, name); found {
		return val, true
	}
	return nil, false
}

// lookupNestedProperty resolves a dotted property path (e.g. "engine_config.galera.enabled")
// within a nested map structure.
func lookupNestedProperty(props biproperty.Map, name string) (interface{}, bool) {
	parts := strings.SplitN(name, ".", 2)
	val, ok := props[parts[0]]
	if !ok {
		return nil, false
	}
	if len(parts) == 1 {
		return val, true
	}
	// Recurse into the nested map.
	switch nested := val.(type) {
	case biproperty.Map:
		return lookupNestedProperty(nested, parts[1])
	case map[string]interface{}:
		converted := make(biproperty.Map, len(nested))
		for k, v := range nested {
			converted[k] = v
		}
		return lookupNestedProperty(converted, parts[1])
	default:
		return nil, false
	}
}

// AllInstanceSpecs constructs the LinkInstanceSpec slice for every instance of a
// deployment job by reading the static_ips from the first network.
// This is the authoritative list passed into the "instances" field of every LinkSpec.
func AllInstanceSpecs(
	deploymentJob bideplmanifest.Job,
	defaultNetworkName string,
) []bitemplate.LinkInstanceSpec {
	instances := make([]bitemplate.LinkInstanceSpec, deploymentJob.Instances)

	var staticIPs []string
	for _, n := range deploymentJob.Networks {
		if n.Name == defaultNetworkName || len(staticIPs) == 0 {
			staticIPs = n.StaticIPs
		}
		if n.Name == defaultNetworkName {
			break
		}
	}

	for i := range instances {
		address := ""
		if i < len(staticIPs) {
			address = staticIPs[i]
		} else if len(staticIPs) > 0 {
			address = staticIPs[0]
		}
		instances[i] = bitemplate.LinkInstanceSpec{
			Name:      deploymentJob.Name,
			ID:        fmt.Sprintf("%s-%d", deploymentJob.Name, i),
			Index:     i,
			Bootstrap: i == 0,
			AZ:        "",
			Address:   address,
		}
	}
	return instances
}
