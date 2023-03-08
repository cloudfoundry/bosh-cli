// This file is part of CycloneDX Go
//
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) OWASP Foundation. All Rights Reserved.

package cyclonedx

import "fmt"

// copyAndConvert returns a converted copy of the BOM, adhering to a given SpecVersion.
func (b BOM) copyAndConvert(specVersion SpecVersion) (*BOM, error) {
	var bomCopy BOM
	err := b.copy(&bomCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to copy bom: %w", err)
	}

	bomCopy.convert(specVersion)
	return &bomCopy, nil
}

// convert modifies the BOM such that it adheres to a given SpecVersion.
func (b *BOM) convert(specVersion SpecVersion) {
	if specVersion < SpecVersion1_1 {
		b.SerialNumber = ""
		b.ExternalReferences = nil
	}
	if specVersion < SpecVersion1_2 {
		b.Dependencies = nil
		b.Metadata = nil
		b.Services = nil
	}
	if specVersion < SpecVersion1_3 {
		b.Compositions = nil
	}
	if specVersion < SpecVersion1_4 {
		b.Vulnerabilities = nil
	}

	if b.Metadata != nil {
		if specVersion < SpecVersion1_3 {
			b.Metadata.Licenses = nil
			b.Metadata.Properties = nil
		}

		recurseComponent(b.Metadata.Component, componentConverter(specVersion))
		convertLicenses(b.Metadata.Licenses, specVersion)
		if b.Metadata.Tools != nil {
			for i := range *b.Metadata.Tools {
				convertTool(&(*b.Metadata.Tools)[i], specVersion)
			}
		}
	}

	if b.Components != nil {
		for i := range *b.Components {
			recurseComponent(&(*b.Components)[i], componentConverter(specVersion))
		}
	}

	if b.Services != nil {
		for i := range *b.Services {
			recurseService(&(*b.Services)[i], serviceConverter(specVersion))
		}
	}

	b.SpecVersion = specVersion
	b.XMLNS = xmlNamespaces[specVersion]
}

// componentConverter modifies a Component such that it adheres to a given SpecVersion.
func componentConverter(specVersion SpecVersion) func(*Component) {
	return func(c *Component) {
		if specVersion < SpecVersion1_1 {
			c.BOMRef = ""
			c.ExternalReferences = nil
			if c.Modified == nil {
				c.Modified = Bool(false)
			}
			c.Pedigree = nil
		}

		if specVersion < SpecVersion1_2 {
			c.Author = ""
			c.MIMEType = ""
			if c.Pedigree != nil {
				c.Pedigree.Patches = nil
			}
			c.Supplier = nil
			c.SWID = nil
		}

		if specVersion < SpecVersion1_3 {
			c.Evidence = nil
			c.Properties = nil
		}

		if specVersion < SpecVersion1_4 {
			c.ReleaseNotes = nil
			if c.Version == "" {
				c.Version = "0.0.0"
			}
		}

		if !specVersion.supportsComponentType(c.Type) {
			c.Type = ComponentTypeApplication
		}
		convertExternalReferences(c.ExternalReferences, specVersion)
		convertHashes(c.Hashes, specVersion)
		convertLicenses(c.Licenses, specVersion)
		if !specVersion.supportsScope(c.Scope) {
			c.Scope = ""
		}
	}
}

// convertExternalReferences modifies an ExternalReference slice such that it adheres to a given SpecVersion.
func convertExternalReferences(extRefs *[]ExternalReference, specVersion SpecVersion) {
	if extRefs == nil {
		return
	}

	if specVersion < SpecVersion1_3 {
		for i := range *extRefs {
			(*extRefs)[i].Hashes = nil
		}
	}
}

// convertHashes modifies a Hash slice such that it adheres to a given SpecVersion.
// If after the conversion no valid hashes are left in the slice, it will be nilled.
func convertHashes(hashes *[]Hash, specVersion SpecVersion) {
	if hashes == nil {
		return
	}

	converted := make([]Hash, 0)
	for i := range *hashes {
		hash := (*hashes)[i]
		if specVersion.supportsHashAlgorithm(hash.Algorithm) {
			converted = append(converted, hash)
		}
	}

	if len(converted) == 0 {
		*hashes = nil
	} else {
		*hashes = converted
	}
}

// convertLicenses modifies a Licenses slice such that it adheres to a given SpecVersion.
// If after the conversion no valid licenses are left in the slice, it will be nilled.
func convertLicenses(licenses *Licenses, specVersion SpecVersion) {
	if licenses == nil {
		return
	}

	if specVersion < SpecVersion1_1 {
		converted := make(Licenses, 0)
		for i := range *licenses {
			choice := &(*licenses)[i]
			if choice.License != nil {
				if choice.License.ID == "" && choice.License.Name == "" {
					choice.License = nil
				} else {
					choice.License.Text = nil
					choice.License.URL = ""
				}
			}
			choice.Expression = ""
			if choice.License != nil {
				converted = append(converted, *choice)
			}
		}

		if len(converted) == 0 {
			*licenses = nil
		} else {
			*licenses = converted
		}
	}
}

// serviceConverter modifies a Service such that it adheres to a given SpecVersion.
func serviceConverter(specVersion SpecVersion) func(*Service) {
	return func(s *Service) {
		if specVersion < SpecVersion1_3 {
			s.Properties = nil
		}

		if specVersion < SpecVersion1_4 {
			s.ReleaseNotes = nil
		}

		convertExternalReferences(s.ExternalReferences, specVersion)
	}
}

// convertTool modifies a Tool such that it adheres to a given SpecVersion.
func convertTool(tool *Tool, specVersion SpecVersion) {
	if tool == nil {
		return
	}

	if specVersion < SpecVersion1_4 {
		tool.ExternalReferences = nil
	}

	convertExternalReferences(tool.ExternalReferences, specVersion)
	convertHashes(tool.Hashes, specVersion)
}

func recurseComponent(component *Component, f func(c *Component)) {
	if component == nil {
		return
	}

	f(component)

	if component.Components != nil {
		for i := range *component.Components {
			recurseComponent(&(*component.Components)[i], f)
		}
	}
	if component.Pedigree != nil {
		if component.Pedigree.Ancestors != nil {
			for i := range *component.Pedigree.Ancestors {
				recurseComponent(&(*component.Pedigree.Ancestors)[i], f)
			}
		}
		if component.Pedigree.Descendants != nil {
			for i := range *component.Pedigree.Descendants {
				recurseComponent(&(*component.Pedigree.Descendants)[i], f)
			}
		}
		if component.Pedigree.Variants != nil {
			for i := range *component.Pedigree.Variants {
				recurseComponent(&(*component.Pedigree.Variants)[i], f)
			}
		}
	}
}

func recurseService(service *Service, f func(s *Service)) {
	if service == nil {
		return
	}

	f(service)

	if service.Services != nil {
		for i := range *service.Services {
			recurseService(&(*service.Services)[i], f)
		}
	}
}

func (sv SpecVersion) supportsComponentType(cType ComponentType) bool {
	switch cType {
	case ComponentTypeApplication, ComponentTypeDevice, ComponentTypeFramework, ComponentTypeLibrary, ComponentTypeOS:
		return sv >= SpecVersion1_0
	case ComponentTypeFile:
		return sv >= SpecVersion1_1
	case ComponentTypeContainer, ComponentTypeFirmware:
		return sv >= SpecVersion1_2
	}

	return false
}

func (sv SpecVersion) supportsHashAlgorithm(algo HashAlgorithm) bool {
	switch algo {
	case HashAlgoMD5, HashAlgoSHA1, HashAlgoSHA256, HashAlgoSHA384, HashAlgoSHA512, HashAlgoSHA3_256, HashAlgoSHA3_512:
		return sv >= SpecVersion1_0
	case HashAlgoSHA3_384, HashAlgoBlake2b_256, HashAlgoBlake2b_384, HashAlgoBlake2b_512, HashAlgoBlake3:
		return sv >= SpecVersion1_2
	}

	return false
}

func (sv SpecVersion) supportsScope(scope Scope) bool {
	switch scope {
	case ScopeRequired, ScopeOptional:
		return sv >= SpecVersion1_0
	case ScopeExcluded:
		return sv >= SpecVersion1_2
	}

	return false
}
