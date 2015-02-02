package erbrenderer

import (
	"strings"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

type PropertiesResolver interface {
	Resolve() bmproperty.Map
}

type propertiesResolver struct {
	defaultProperties    bmproperty.Map
	propertiesWithValues bmproperty.Map
}

func NewPropertiesResolver(
	defaultProperties bmproperty.Map,
	propertiesWithValues bmproperty.Map,
) PropertiesResolver {
	return propertiesResolver{
		defaultProperties:    defaultProperties,
		propertiesWithValues: propertiesWithValues,
	}
}

func (p propertiesResolver) Resolve() bmproperty.Map {
	result := p.propertiesWithValues
	for propertyKey, defaultPropertyValue := range p.defaultProperties {
		propertyKeyParts := strings.Split(propertyKey, ".")
		p.copyDefault(result, propertyKeyParts, defaultPropertyValue)
	}

	return result
}

func (p propertiesResolver) copyDefault(valuesMap bmproperty.Map, keyPath []string, defaultValue bmproperty.Property) {
	keyName := keyPath[0]

	if len(keyPath) == 1 {
		if _, ok := valuesMap[keyName]; !ok {
			valuesMap[keyName] = defaultValue
		}
		return
	}

	var innerValuesMap bmproperty.Map

	if innerValues, ok := valuesMap[keyName]; ok {
		if innerValuesMap, ok = innerValues.(bmproperty.Map); !ok {
			// Value is set already
			return
		}
	} else {
		// Value is missing, initialize map for the default value
		innerValuesMap = bmproperty.Map{}
		valuesMap[keyName] = innerValuesMap
	}

	p.copyDefault(innerValuesMap, keyPath[1:], defaultValue)
}
