package releasedir

import (
	"fmt"
	"sort"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

/*
# final.yml
---
name: cf
blobstore:
  provider: s3
  options:
    bucket_name: cf-release-blobs

# private.yml
---
blobstore:
  options: { ... }
*/

type FSConfig struct {
	publicPath  string
	privatePath string
	fs          boshsys.FileSystem
}

type fsConfigPublicSchema struct {
	Name          string                   `yaml:"name"`
	FinalName     string                   `yaml:"final_name,omitempty"`
	Blobstore     fsConfigSchema_Blobstore `yaml:"blobstore,omitempty"`
	NoCompression bool                     `yaml:"no_compression,omitempty"`
}

type fsConfigPrivateSchema struct {
	Blobstore fsConfigSchema_Blobstore `yaml:"blobstore"`
}

type fsConfigSchema_Blobstore struct {
	Provider string                 `yaml:"provider"`
	Options  map[string]interface{} `yaml:"options,omitempty"`
}

func NewFSConfig(publicPath, privatePath string, fs boshsys.FileSystem) FSConfig {
	return FSConfig{publicPath: publicPath, privatePath: privatePath, fs: fs}
}

func (c FSConfig) Name() (string, error) {
	publicSchema, _, err := c.read()
	if err != nil {
		return "", err
	}

	if len(publicSchema.Name) == 0 {
		if len(publicSchema.FinalName) == 0 {
			return "", bosherr.Errorf(
				"Expected non-empty 'name' in config '%s'", c.publicPath)
		}

		return publicSchema.FinalName, nil
	} else if len(publicSchema.FinalName) > 0 {
		return "", bosherr.Errorf(
			"Expected 'name' or 'final_name' but not both in config '%s'", c.publicPath)
	}

	return publicSchema.Name, nil
}

func (c FSConfig) SaveName(name string) error {
	schema, err := c.readMap(c.publicPath)
	if err != nil {
		return err
	}

	schema = removeMapItem(schema, "final_name")
	schema = setMapItem(schema, "name", name)

	return c.writeMap(c.publicPath, schema)
}

func (c FSConfig) SaveBlobstore(provider string, options, privateOptions map[string]interface{}) error {
	publicSchema, err := c.readMap(c.publicPath)
	if err != nil {
		return err
	}

	blobstore := mapItemMapSlice(publicSchema, "blobstore")

	existingProvider, _ := mapItemValue(blobstore, "provider").(string)
	providerChanged := existingProvider != provider

	existingOptions := mapItemValue(blobstore, "options")
	if providerChanged {
		existingOptions = nil
	}

	blobstore = setMapItem(blobstore, "provider", provider)

	mergedOptions := mergeOptions(existingOptions, options)

	if provider == "local" {
		if _, found := mergedOptions["blobstore_path"]; !found {
			mergedOptions["blobstore_path"] = ".blobs"
		}
	}

	blobstore = setBlobstoreOptions(blobstore, mergedOptions)

	publicSchema = setMapItem(publicSchema, "blobstore", blobstore)

	err = c.writeMap(c.publicPath, publicSchema)
	if err != nil {
		return err
	}

	if len(privateOptions) == 0 && !providerChanged {
		return nil
	}

	if !c.fs.FileExists(c.privatePath) && len(privateOptions) == 0 {
		return nil
	}

	privateSchema, err := c.readMap(c.privatePath)
	if err != nil {
		return err
	}

	privateBlobstore := mapItemMapSlice(privateSchema, "blobstore")

	privateExistingOptions := mapItemValue(privateBlobstore, "options")
	if providerChanged {
		privateExistingOptions = nil
	}

	privateBlobstore = setBlobstoreOptions(privateBlobstore, mergeOptions(privateExistingOptions, privateOptions))
	privateSchema = setMapItem(privateSchema, "blobstore", privateBlobstore)

	return c.writeMap(c.privatePath, privateSchema)
}

func setBlobstoreOptions(blobstore yaml.MapSlice, options map[interface{}]interface{}) yaml.MapSlice {
	if len(options) == 0 {
		return removeMapItem(blobstore, "options")
	}

	return setMapItem(blobstore, "options", toMapSlice(options))
}

func (c FSConfig) readMap(path string) (yaml.MapSlice, error) {
	if !c.fs.FileExists(path) {
		return yaml.MapSlice{}, nil
	}

	bytes, err := c.fs.ReadFile(path)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading config '%s'", path)
	}

	var raw interface{}

	err = yaml.Unmarshal(bytes, &raw)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Unmarshalling config '%s'", path)
	}

	switch raw.(type) {
	case nil, map[interface{}]interface{}:
	default:
		return nil, bosherr.Errorf("Unmarshalling config '%s': expected a mapping", path)
	}

	var schema yaml.MapSlice

	err = yaml.Unmarshal(bytes, &schema)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Unmarshalling config '%s'", path)
	}

	return schema, nil
}

func (c FSConfig) writeMap(path string, schema yaml.MapSlice) error {
	bytes, err := yaml.Marshal(schema)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshalling config '%s'", path)
	}

	err = c.fs.WriteFile(path, bytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing config '%s'", path)
	}

	return nil
}

func setMapItem(schema yaml.MapSlice, key string, value interface{}) yaml.MapSlice {
	for i, item := range schema {
		if item.Key == key {
			schema[i].Value = value
			return schema
		}
	}

	return append(schema, yaml.MapItem{Key: key, Value: value})
}

func removeMapItem(schema yaml.MapSlice, key string) yaml.MapSlice {
	result := make(yaml.MapSlice, 0, len(schema))

	for _, item := range schema {
		if item.Key != key {
			result = append(result, item)
		}
	}

	return result
}

func mapItemValue(schema yaml.MapSlice, key string) interface{} {
	for _, item := range schema {
		if item.Key == key {
			return item.Value
		}
	}

	return nil
}

func mapItemMapSlice(schema yaml.MapSlice, key string) yaml.MapSlice {
	for _, item := range schema {
		if item.Key != key {
			continue
		}

		switch value := item.Value.(type) {
		case yaml.MapSlice:
			return value
		case map[interface{}]interface{}:
			return toMapSlice(value)
		}
	}

	return yaml.MapSlice{}
}

func toMapSlice(values map[interface{}]interface{}) yaml.MapSlice {
	keys := make([]string, 0, len(values))

	for key := range values {
		keys = append(keys, fmt.Sprintf("%v", key))
	}

	sort.Strings(keys)

	schema := make(yaml.MapSlice, 0, len(values))

	for _, key := range keys {
		schema = append(schema, yaml.MapItem{Key: key, Value: values[key]})
	}

	return schema
}

func mergeOptions(existing interface{}, additional map[string]interface{}) map[interface{}]interface{} {
	result := map[interface{}]interface{}{}

	switch options := existing.(type) {
	case map[interface{}]interface{}:
		for key, value := range options {
			result[key] = value
		}
	case yaml.MapSlice:
		for _, item := range options {
			result[item.Key] = item.Value
		}
	}

	for key, value := range additional {
		result[key] = value
	}

	return result
}

func (c FSConfig) NoCompression() bool {
	publicSchema, _, err := c.read()
	if err != nil {
		return false
	}
	return publicSchema.NoCompression
}

func (c FSConfig) Blobstore() (string, map[string]interface{}, error) {
	publicSchema, privateSchema, err := c.read()
	if err != nil {
		return "", nil, err
	}

	if len(publicSchema.Blobstore.Provider) == 0 {
		return "", nil, bosherr.Errorf(
			"Expected non-empty 'blobstore.provider' in config '%s'", c.publicPath)
	}

	opts := map[string]interface{}{}

	for k, v := range publicSchema.Blobstore.Options {
		opts[k] = v
	}

	for k, v := range privateSchema.Blobstore.Options {
		opts[k] = v
	}

	return publicSchema.Blobstore.Provider, opts, nil
}

func (c FSConfig) read() (fsConfigPublicSchema, fsConfigPrivateSchema, error) {
	var publicSchema fsConfigPublicSchema
	var privateSchema fsConfigPrivateSchema

	if c.fs.FileExists(c.publicPath) {
		bytes, err := c.fs.ReadFile(c.publicPath)
		if err != nil {
			return publicSchema, privateSchema,
				bosherr.WrapErrorf(err, "Reading config '%s'", c.publicPath)
		}

		err = yaml.Unmarshal(bytes, &publicSchema)
		if err != nil {
			return publicSchema, privateSchema,
				bosherr.WrapErrorf(err, "Unmarshalling config '%s'", c.publicPath)
		}
	}

	if c.fs.FileExists(c.privatePath) {
		bytes, err := c.fs.ReadFile(c.privatePath)
		if err != nil {
			return publicSchema, privateSchema,
				bosherr.WrapErrorf(err, "Reading config '%s'", c.privatePath)
		}

		err = yaml.Unmarshal(bytes, &privateSchema)
		if err != nil {
			return publicSchema, privateSchema,
				bosherr.WrapErrorf(err, "Unmarshalling config '%s'", c.privatePath)
		}
	}

	return publicSchema, privateSchema, nil
}
