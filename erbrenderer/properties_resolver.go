package erbrenderer

import (
	"strings"

	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/jobs"
)

type PropertiesResolver interface {
	Resolve() map[string]interface{}
}

type propertiesResolver struct {
	defaultProperties    map[string]bmreljob.PropertyDefinition
	propertiesWithValues map[string]interface{}
}

func NewPropertiesResolver(
	defaultProperties map[string]bmreljob.PropertyDefinition,
	propertiesWithValues map[string]interface{},
) PropertiesResolver {
	return propertiesResolver{
		defaultProperties:    defaultProperties,
		propertiesWithValues: propertiesWithValues,
	}
}

func (p propertiesResolver) Resolve() map[string]interface{} {
	result := p.propertiesWithValues
	for propertyKey, defaultProperty := range p.defaultProperties {
		propertyKeyParts := strings.Split(propertyKey, ".")
		p.copyDefault(result, propertyKeyParts, defaultProperty.Default)
	}

	return result
}

func (p propertiesResolver) copyDefault(valuesMap map[string]interface{}, keyPath []string, defaultValue interface{}) {
	keyName := keyPath[0]

	if len(keyPath) == 1 {
		if _, ok := valuesMap[keyName]; !ok {
			valuesMap[keyName] = defaultValue
		}
		return
	}

	var innerValuesMap map[string]interface{}

	if innerValues, ok := valuesMap[keyName]; ok {
		if innerValuesMap, ok = innerValues.(map[string]interface{}); !ok {
			// Value is set already
			return
		}
	} else {
		// Value is missing, initialize map for the default value
		innerValuesMap = map[string]interface{}{}
		valuesMap[keyName] = innerValuesMap
	}

	p.copyDefault(innerValuesMap, keyPath[1:], defaultValue)
}
