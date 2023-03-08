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

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

// bomReferenceXML is temporarily used for marshalling and unmarshalling
// BOMReference instances to and from XML.
type bomReferenceXML struct {
	Ref string `json:"-" xml:"ref,attr"`
}

func (b BOMReference) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(bomReferenceXML{Ref: string(b)}, start)
}

func (b *BOMReference) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	bXML := bomReferenceXML{}
	if err := d.DecodeElement(&bXML, &start); err != nil {
		return err
	}
	*b = BOMReference(bXML.Ref)
	return nil
}

func (c Copyright) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(c.Text, start)
}

func (c *Copyright) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var text string
	if err := d.DecodeElement(&text, &start); err != nil {
		return err
	}
	(*c).Text = text
	return nil
}

// dependencyXML is temporarily used for marshalling and unmarshalling
// Dependency instances to and from XML.
type dependencyXML struct {
	Ref          string           `xml:"ref,attr"`
	Dependencies *[]dependencyXML `xml:"dependency,omitempty"`
}

func (d Dependency) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	xmlDep := dependencyXML{Ref: d.Ref}

	if d.Dependencies != nil && len(*d.Dependencies) > 0 {
		xmlDeps := make([]dependencyXML, len(*d.Dependencies))
		for i := range *d.Dependencies {
			xmlDeps[i] = dependencyXML{Ref: (*d.Dependencies)[i]}
		}
		xmlDep.Dependencies = &xmlDeps
	}

	return e.EncodeElement(xmlDep, start)
}

func (d *Dependency) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	xmlDep := dependencyXML{}
	err := dec.DecodeElement(&xmlDep, &start)
	if err != nil {
		return err
	}

	dep := Dependency{Ref: xmlDep.Ref}
	if xmlDep.Dependencies != nil && len(*xmlDep.Dependencies) > 0 {
		deps := make([]string, len(*xmlDep.Dependencies))
		for i := range *xmlDep.Dependencies {
			deps[i] = (*xmlDep.Dependencies)[i].Ref
		}
		dep.Dependencies = &deps
	}

	*d = dep
	return nil
}

func (l Licenses) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(l) == 0 {
		return nil
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for _, choice := range l {
		if choice.License != nil && choice.Expression != "" {
			return fmt.Errorf("either license or expression must be set, but not both")
		}

		if choice.License != nil {
			if err := e.EncodeElement(choice.License, xml.StartElement{Name: xml.Name{Local: "license"}}); err != nil {
				return err
			}
		} else if choice.Expression != "" {
			if err := e.EncodeElement(choice.Expression, xml.StartElement{Name: xml.Name{Local: "expression"}}); err != nil {
				return err
			}
		}
	}

	return e.EncodeToken(start.End())
}

func (l *Licenses) UnmarshalXML(d *xml.Decoder, _ xml.StartElement) error {
	licenses := make([]LicenseChoice, 0)

	for {
		token, err := d.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		switch tokenType := token.(type) {
		case xml.StartElement:
			if tokenType.Name.Local == "expression" {
				var expression string
				if err = d.DecodeElement(&expression, &tokenType); err != nil {
					return err
				}
				licenses = append(licenses, LicenseChoice{Expression: expression})
			} else if tokenType.Name.Local == "license" {
				var license License
				if err = d.DecodeElement(&license, &tokenType); err != nil {
					return err
				}
				licenses = append(licenses, LicenseChoice{License: &license})
			} else {
				return fmt.Errorf("unknown element: %s", tokenType.Name.Local)
			}
		}
	}

	*l = licenses
	return nil
}

func (sv SpecVersion) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(sv.String(), start)
}

func (sv *SpecVersion) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	err := d.DecodeElement(&v, &start)
	if err != nil {
		return err
	}

	switch v {
	case SpecVersion1_0.String():
		*sv = SpecVersion1_0
	case SpecVersion1_1.String():
		*sv = SpecVersion1_1
	case SpecVersion1_2.String():
		*sv = SpecVersion1_2
	case SpecVersion1_3.String():
		*sv = SpecVersion1_3
	case SpecVersion1_4.String():
		*sv = SpecVersion1_4
	default:
		return ErrInvalidSpecVersion
	}

	return nil
}

var xmlNamespaces = map[SpecVersion]string{
	SpecVersion1_0: "http://cyclonedx.org/schema/bom/1.0",
	SpecVersion1_1: "http://cyclonedx.org/schema/bom/1.1",
	SpecVersion1_2: "http://cyclonedx.org/schema/bom/1.2",
	SpecVersion1_3: "http://cyclonedx.org/schema/bom/1.3",
	SpecVersion1_4: "http://cyclonedx.org/schema/bom/1.4",
}
