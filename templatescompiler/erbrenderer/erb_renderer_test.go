package erbrenderer_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/templatescompiler/erbrenderer"
	"github.com/cloudfoundry/bosh-cli/v7/templatescompiler/erbrenderer/erbrendererfakes"
)

type testTemplateEvaluationStruct struct {
	Index             int                    `json:"index"`
	ID                string                 `json:"id"`
	IP                string                 `json:"ip,omitempty"`
	GlobalProperties  map[string]interface{} `json:"global_properties"`
	ClusterProperties map[string]interface{} `json:"cluster_properties"`
	DefaultProperties map[string]interface{} `json:"default_properties"`
}

type testTemplateEvaluationContext struct {
	testStruct testTemplateEvaluationStruct
}

func (t testTemplateEvaluationContext) MarshalJSON() ([]byte, error) {
	jsonBytes, err := json.Marshal(t.testStruct)
	if err != nil {
		return []byte{}, err
	}

	return jsonBytes, nil
}

var _ = Describe("ErbRenderer", func() {
	Describe("Render", func() {
		var (
			fs     boshsys.FileSystem
			runner boshsys.CmdRunner
			logger boshlog.Logger

			tmpDir                   string
			erbTemplateFilepath      string
			renderedTemplatePath     string
			erbTemplateContent       string
			expectedTemplateContents string

			context erbrenderer.TemplateEvaluationContext

			erbRenderer erbrenderer.ERBRenderer
		)

		BeforeEach(func() {
			logger = boshlog.NewLogger(boshlog.LevelNone)
			fs = boshsys.NewOsFileSystemWithStrictTempRoot(logger)
			runner = boshsys.NewExecCmdRunner(logger)

			tmpDir = GinkgoT().TempDir()
			Expect(fs.ChangeTempRoot(tmpDir)).To(Succeed())

			templateName := "test_template.yml"
			erbTemplateName := fmt.Sprintf("%s.erb", templateName)
			erbTemplateFilepath = filepath.Join(tmpDir, erbTemplateName)
			renderedTemplatePath = filepath.Join(tmpDir, templateName)

			context = &testTemplateEvaluationContext{
				testTemplateEvaluationStruct{
					Index: 867_5309,
					GlobalProperties: map[string]interface{}{
						"property1": "global_value1",
					},
					ClusterProperties: map[string]interface{}{
						"property2": "cluster_value1",
					},
					DefaultProperties: map[string]interface{}{
						"property1": "default_value1",
						"property2": "default_value2",
						"property3": "default_value3",
					},
				},
			}
		})

		Context("when actually executing `ruby`", func() {
			BeforeEach(func() {
				erbRenderer = erbrenderer.NewERBRenderer(fs, runner, logger)
			})

			Context("with valid ERB", func() {
				BeforeEach(func() {
					erbTemplateContent = `---
property1: <%= p('property1') %>
property2: <%= p('property2') %>
property3: <%= p('property3') %>
`
					expectedTemplateContents = `---
property1: global_value1
property2: cluster_value1
property3: default_value3
`
					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())
				})

				It("renders output", func() {
					err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())
					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())

					Expect(expectedTemplateContents).To(Equal(string(templateBytes)))
				})
			})

			Context("with nested property access patterns", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"director": map[string]interface{}{
									"db": map[string]interface{}{
										"user":     "admin",
										"password": "secret",
										"host":     "localhost",
										"port":     5432,
									},
									"name": "test-director",
								},
								"nats": map[string]interface{}{
									"address": "10.0.0.1",
									"port":    4222,
								},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"director.db.user":     "default_user",
								"director.db.password": "default_pass",
								"director.db.host":     "default_host",
								"director.db.port":     "default_port",
								"director.name":        "default_name",
								"nats.address":         "default_nats",
								"nats.port":            "default_port",
							},
						},
					}
				})

				It("accesses deeply nested properties with dot notation", func() {
					erbTemplateContent = `db_user: <%= p('director.db.user') %>
db_pass: <%= p('director.db.password') %>
db_host: <%= p('director.db.host') %>
nats_addr: <%= p('nats.address') %>`
					expectedTemplateContents = `db_user: admin
db_pass: secret
db_host: localhost
nats_addr: 10.0.0.1`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with property defaults", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index:             0,
							GlobalProperties:  map[string]interface{}{},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"existing_prop": "default_value",
							},
						},
					}
				})

				It("uses default values for missing properties", func() {
					erbTemplateContent = `has_default: <%= p('missing_prop', 'fallback_value') %>
no_default: <%= p('existing_prop') %>`
					expectedTemplateContents = `has_default: fallback_value
no_default: default_value`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with if_p conditional property helper", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"enabled_feature": true,
								"feature_config": map[string]interface{}{
									"host": "example.com",
									"port": 8080,
								},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"enabled_feature":      "default",
								"feature_config.host":  "default",
								"feature_config.port":  "default",
								"missing_feature.host": "default",
							},
						},
					}
				})

				It("executes block when property exists", func() {
					erbTemplateContent = `<% if_p('feature_config.host') do |host| %>
host_configured: <%= host %>
<% end -%>`
					expectedTemplateContents = `
host_configured: example.com
`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})

				It("supports multiple properties in if_p", func() {
					erbTemplateContent = `<% if_p('feature_config.host', 'feature_config.port') do |host, port| %>
config: <%= host %>:<%= port %>
<% end -%>`
					expectedTemplateContents = `
config: example.com:8080
`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})

				It("skips block when property is missing", func() {
					erbTemplateContent = `before
<% if_p('completely_missing_prop') do |host| %>
should_not_appear: <%= host %>
<% end -%>
after`
					expectedTemplateContents = `before
after`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with array of hashes property access", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"users": []interface{}{
									map[string]interface{}{
										"name":     "alice",
										"password": "secret1",
									},
									map[string]interface{}{
										"name":     "bob",
										"password": "secret2",
									},
								},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"users": "default",
							},
						},
					}
				})

				It("iterates over array and accesses hash elements", func() {
					erbTemplateContent = `<% p('users').each do |user| %>
user: <%= user['name'] %> pass: <%= user['password'] %>
<% end -%>`
					expectedTemplateContents = `
user: alice pass: secret1

user: bob pass: secret2
`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with spec object access", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index:             42,
							ID:                "uuid-123-456",
							IP:                "192.168.1.100",
							GlobalProperties:  map[string]interface{}{},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{},
						},
					}
				})

				It("accesses spec properties via struct notation", func() {
					erbTemplateContent = `index: <%= spec.index %>
id: <%= spec.id %>
ip: <%= spec.ip %>`
					expectedTemplateContents = `index: 42
id: uuid-123-456
ip: 192.168.1.100`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with complex nested object creation", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"blobstore": map[string]interface{}{
									"provider": "s3",
									"s3": map[string]interface{}{
										"bucket":     "my-bucket",
										"access_key": "AKIAIOSFODNN7EXAMPLE",
									},
								},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"blobstore.provider":      "default",
								"blobstore.s3.bucket":     "default",
								"blobstore.s3.access_key": "default",
							},
						},
					}
				})

				It("builds nested hash structures from properties", func() {
					erbTemplateContent = `<%=
config = {
  'provider' => p('blobstore.provider'),
  'options' => {
    'bucket' => p('blobstore.s3.bucket'),
    'access_key' => p('blobstore.s3.access_key')
  }
}
require 'json'
JSON.dump(config)
%>`
					expectedTemplateContents = `{"provider":"s3","options":{"bucket":"my-bucket","access_key":"AKIAIOSFODNN7EXAMPLE"}}`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with boolean and numeric property types", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"features": map[string]interface{}{
									"enabled":    true,
									"max_count":  100,
									"timeout":    30.5,
									"debug_mode": false,
								},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"features.enabled":    "default",
								"features.max_count":  "default",
								"features.timeout":    "default",
								"features.debug_mode": "default",
							},
						},
					}
				})

				It("handles boolean and numeric property values", func() {
					erbTemplateContent = `enabled: <%= p('features.enabled') %>
max_count: <%= p('features.max_count') %>
timeout: <%= p('features.timeout') %>
debug: <%= p('features.debug_mode') %>`
					expectedTemplateContents = `enabled: true
max_count: 100
timeout: 30.5
debug: false`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})

				It("uses booleans in conditionals", func() {
					erbTemplateContent = `<% if p('features.enabled') -%>
feature is enabled
<% end -%>
<% if !p('features.debug_mode') -%>
debug is disabled
<% end -%>`
					expectedTemplateContents = `feature is enabled
debug is disabled
`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with array map operations", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"ports": []interface{}{"8080", "8443", "9000"},
								"servers": []interface{}{
									map[string]interface{}{"host": "10.0.0.1", "port": 8080},
									map[string]interface{}{"host": "10.0.0.2", "port": 8081},
								},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"ports":   "default",
								"servers": "default",
							},
						},
					}
				})

				It("converts array elements using map with symbol", func() {
					erbTemplateContent = `<%= p('ports').map(&:to_i).inspect %>`
					expectedTemplateContents = `[8080, 8443, 9000]`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})

				It("transforms array elements using map with block", func() {
					erbTemplateContent = `<% p('servers').map { |s| s['host'] }.each do |host| %>
host: <%= host %>
<% end -%>`
					expectedTemplateContents = `
host: 10.0.0.1

host: 10.0.0.2
`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with array filtering operations", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"ca_certs": []interface{}{
									"",
									"  ",
									nil,
									"-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----",
									"short",
									"-----BEGIN CERTIFICATE-----\nMIID...\n-----END CERTIFICATE-----",
								},
								"ports": []interface{}{8080, nil, 8443, nil, 9000},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"ca_certs": "default",
								"ports":    "default",
							},
						},
					}
				})

				It("filters array elements using select", func() {
					erbTemplateContent = `<%= p('ca_certs').select{ |v| !v.nil? && !v.strip.empty? && v.length > 50 }.length %>`
					expectedTemplateContents = `2`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})

				It("removes nil values using compact", func() {
					erbTemplateContent = `<%= p('ports').compact.inspect %>`
					expectedTemplateContents = `[8080, 8443, 9000]`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with method chaining on property values", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"multiline_cert": "-----BEGIN CERTIFICATE-----\nline1\nline2\n-----END CERTIFICATE-----\n",
								"url":            "https://example.com:8443/path",
								"yaml_data": map[string]interface{}{
									"key": "value",
								},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"multiline_cert": "default",
								"url":            "default",
								"yaml_data":      "default",
							},
						},
					}
				})

				It("chains methods on string properties", func() {
					erbTemplateContent = `<%= p('url').split(':')[0] %>`
					expectedTemplateContents = `https`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})

				It("processes multiline strings with lines and map", func() {
					erbTemplateContent = `<%= p('multiline_cert').lines.map { |line| "  #{line.rstrip}" }.join("\n") %>`
					expectedTemplateContents = `  -----BEGIN CERTIFICATE-----
  line1
  line2
  -----END CERTIFICATE-----`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})

				It("chains to_yaml and gsub on hash properties", func() {
					erbTemplateContent = `<%= p('yaml_data').to_yaml.gsub("---","").strip %>`
					expectedTemplateContents = `key: value`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with each_with_index iteration", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"routes": []interface{}{
									map[string]interface{}{"uri": "api.example.com"},
									map[string]interface{}{"uri": "www.example.com"},
								},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"routes": "default",
							},
						},
					}
				})

				It("iterates with index access", func() {
					erbTemplateContent = `<% p('routes').each_with_index do |route, index| %>
route_<%= index %>: <%= route['uri'] %>
<% end -%>`
					expectedTemplateContents = `
route_0: api.example.com

route_1: www.example.com
`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with hash key access and membership testing", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"clients": map[string]interface{}{
									"client_a": map[string]interface{}{"secret": "secret_a"},
									"client_b": map[string]interface{}{"secret": "secret_b"},
								},
								"config": map[string]interface{}{
									"optional_key": "value",
									"required_key": "required",
								},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"clients": "default",
								"config":  "default",
							},
						},
					}
				})

				It("accesses hash keys and sorts them", func() {
					erbTemplateContent = `<%= p('clients').keys.sort.first %>`
					expectedTemplateContents = `client_a`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})

				It("checks for hash key membership", func() {
					erbTemplateContent = `<% if p('config').key?('optional_key') %>
has_key: true
<% end -%>`
					expectedTemplateContents = `
has_key: true
`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with conditional string operations", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"api_url":  "https://api.example.com",
								"cert":     "",
								"endpoint": "routing-api.service.cf.internal",
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"api_url":  "default",
								"cert":     "default",
								"endpoint": "default",
							},
						},
					}
				})

				It("checks string prefix with start_with?", func() {
					erbTemplateContent = `<% if p('api_url').start_with?('https') %>
secure: true
<% end -%>`
					expectedTemplateContents = `
secure: true
`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})

				It("checks for empty strings", func() {
					erbTemplateContent = `<% if p('cert') == "" %>
no_cert: true
<% end -%>`
					expectedTemplateContents = `
no_cert: true
`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})

				It("performs string replacement with gsub", func() {
					erbTemplateContent = `<%= p('endpoint').gsub('.internal', '.external') %>`
					expectedTemplateContents = `routing-api.service.cf.external`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with array find operation", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"databases": []interface{}{
									map[string]interface{}{"tag": "uaa", "name": "uaadb"},
									map[string]interface{}{"tag": "admin", "name": "postgres"},
								},
								"providers": []interface{}{
									map[string]interface{}{"type": "internal", "name": "default"},
									map[string]interface{}{"type": "hsm", "name": "thales"},
								},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"databases": "default",
								"providers": "default",
							},
						},
					}
				})

				It("finds elements in array by condition", func() {
					erbTemplateContent = `<% db = p('databases').find { |d| d['tag'] == 'uaa' } %>
db_name: <%= db['name'] %>
<% provider = p('providers').find { |p| p['type'] == 'hsm' } %>
provider_name: <%= provider['name'] %>`
					expectedTemplateContents = `
db_name: uaadb

provider_name: thales`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with array flatten operation", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"nested_providers": []interface{}{
									[]interface{}{
										map[string]interface{}{"type": "internal"},
										map[string]interface{}{"type": "hsm"},
									},
									[]interface{}{
										map[string]interface{}{"type": "kms-plugin"},
									},
								},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"nested_providers": "default",
							},
						},
					}
				})

				It("flattens nested arrays", func() {
					erbTemplateContent = `<%= p('nested_providers').flatten.length %>`
					expectedTemplateContents = `3`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with any? predicate", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"providers": []interface{}{
									map[string]interface{}{"type": "internal"},
									map[string]interface{}{"type": "hsm"},
								},
								"empty_list": []interface{}{},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"providers":  "default",
								"empty_list": "default",
							},
						},
					}
				})

				It("checks if any element matches condition", func() {
					erbTemplateContent = `<% if p('providers').any? { |p| p['type'] == 'hsm' } -%>
using_hsm: true
<% end -%>
<% if !p('empty_list').any? -%>
list_empty: true
<% end -%>`
					expectedTemplateContents = `using_hsm: true
list_empty: true
`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with nil? and empty? checks", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"optional_cert": "",
								"required_key":  "actual_value",
								"empty_array":   []interface{}{},
								"filled_array":  []interface{}{"item"},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"optional_cert": "default",
								"required_key":  "default",
								"empty_array":   "default",
								"filled_array":  "default",
							},
						},
					}
				})

				It("checks for empty strings and arrays", func() {
					erbTemplateContent = `<% if p('optional_cert').empty? -%>
no_cert: true
<% end -%>
<% if !p('required_key').empty? -%>
has_key: true
<% end -%>
<% if p('empty_array').empty? -%>
array_empty: true
<% end -%>
<% if !p('filled_array').empty? -%>
array_filled: true
<% end -%>`
					expectedTemplateContents = `no_cert: true
has_key: true
array_empty: true
array_filled: true
`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with include? membership testing", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"valid_modes":   []interface{}{"legacy", "exact"},
								"selected_mode": "exact",
								"tls_modes":     []interface{}{"enabled", "disabled"},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"valid_modes":   "default",
								"selected_mode": "default",
								"tls_modes":     "default",
							},
						},
					}
				})

				It("checks array membership", func() {
					erbTemplateContent = `<% if p('valid_modes').include?(p('selected_mode')) -%>
valid_selection: true
<% end -%>
<% if p('tls_modes').include?('enabled') -%>
supports_tls: true
<% end -%>`
					expectedTemplateContents = `valid_selection: true
supports_tls: true
`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with reject and uniq operations", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"providers": []interface{}{
									map[string]interface{}{"name": "p1", "enabled": true},
									map[string]interface{}{"name": "p2", "enabled": false},
									map[string]interface{}{"name": "p3", "enabled": true},
								},
								"types": []interface{}{"internal", "hsm", "internal", "kms-plugin"},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"providers": "default",
								"types":     "default",
							},
						},
					}
				})

				It("rejects unwanted elements", func() {
					erbTemplateContent = `<%= p('providers').reject { |p| !p['enabled'] }.length %>`
					expectedTemplateContents = `2`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})

				It("removes duplicate values", func() {
					erbTemplateContent = `<%= p('types').uniq.inspect %>`
					expectedTemplateContents = `["internal", "hsm", "kms-plugin"]`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with hash values and merge operations", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"config": map[string]interface{}{
									"host":    "localhost",
									"port":    5432,
									"timeout": 30,
								},
								"defaults": map[string]interface{}{
									"timeout": 60,
									"retries": 3,
								},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"config":   "default",
								"defaults": "default",
							},
						},
					}
				})

				It("extracts hash values", func() {
					erbTemplateContent = `<%= p('config').values.sort_by(&:to_s).inspect %>`
					expectedTemplateContents = `[30, 5432, "localhost"]`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})

				It("merges hashes", func() {
					erbTemplateContent = `<% merged = p('defaults').merge(p('config')) %>
timeout: <%= merged['timeout'] %>
retries: <%= merged['retries'] %>`
					expectedTemplateContents = `
timeout: 30
retries: 3`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with string index and type conversion operations", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"cert_with_newlines":    "-----BEGIN CERTIFICATE-----\nMIIC...",
								"cert_without_newlines": "-----BEGIN CERTIFICATE-----MIIC...",
								"port_string":           "8443",
								"timeout_number":        30,
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"cert_with_newlines":    "default",
								"cert_without_newlines": "default",
								"port_string":           "default",
								"timeout_number":        "default",
							},
						},
					}
				})

				It("finds substring positions with index", func() {
					erbTemplateContent = `<% if p('cert_with_newlines').index("\n").nil? %>
no_real_newline: true
<% else %>
has_real_newline: true
<% end -%>
<% if p('cert_without_newlines').index("\n").nil? %>
no_escaped_newline: true
<% end -%>`
					expectedTemplateContents = `
has_real_newline: true

no_escaped_newline: true
`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})

				It("converts types with to_i and to_s", func() {
					erbTemplateContent = `port_number: <%= p('port_string').to_i %>
timeout_string: <%= p('timeout_number').to_s %>`
					expectedTemplateContents = `port_number: 8443
timeout_string: 30`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with first and last array accessors", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"servers": []interface{}{
									"server1.example.com",
									"server2.example.com",
									"server3.example.com",
								},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"servers": "default",
							},
						},
					}
				})

				It("accesses first and last array elements", func() {
					erbTemplateContent = `primary: <%= p('servers').first %>
backup: <%= p('servers').last %>`
					expectedTemplateContents = `primary: server1.example.com
backup: server3.example.com`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Context("with join operation on arrays", func() {
				BeforeEach(func() {
					context = &testTemplateEvaluationContext{
						testTemplateEvaluationStruct{
							Index: 0,
							GlobalProperties: map[string]interface{}{
								"ciphers": []interface{}{
									"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
									"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
								},
								"scopes": []interface{}{"openid", "profile", "email"},
							},
							ClusterProperties: map[string]interface{}{},
							DefaultProperties: map[string]interface{}{
								"ciphers": "default",
								"scopes":  "default",
							},
						},
					}
				})

				It("joins array elements with delimiter", func() {
					erbTemplateContent = `ciphers: <%= p('ciphers').join(',') %>
scopes: <%= p('scopes').join(' ') %>`
					expectedTemplateContents = `ciphers: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
scopes: openid profile email`

					err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
					Expect(err).ToNot(HaveOccurred())

					err = erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).ToNot(HaveOccurred())

					templateBytes, err := fs.ReadFile(renderedTemplatePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(templateBytes)).To(Equal(expectedTemplateContents))
				})
			})

			Describe("error handling within Ruby", func() {
				var (
					// see erb_renderer.rb
					rubyExceptionPrefixTemplate = "Error filling in template '%s' "
				)

				Context("with invalid ERB", func() {
					BeforeEach(func() {
						erbTemplateContent = `<%= raise "test error" %>`

						err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
						Expect(err).ToNot(HaveOccurred())
					})

					It("returns an error with a known ruby exception", func() {
						err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("RuntimeError: test error"))
						Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(rubyExceptionPrefixTemplate, erbTemplateFilepath)))
					})

					Context("with missing ERB template", func() {
						It("returns an error with a known ruby exception", func() {
							invalidErbPath := "invalid/template.erb"
							err := erbRenderer.Render(invalidErbPath, renderedTemplatePath, context)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("<Errno::ENOENT: No such file or directory @ rb_sysopen - %s>", invalidErbPath)))
							Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(rubyExceptionPrefixTemplate, invalidErbPath)))
						})
					})

					Context("with context JSON which does not have the expected elements", func() {
						It("returns an error with a known ruby exception", func() {
							err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, &testTemplateEvaluationContext{})
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("NoMethodError"))
							Expect(err.Error()).To(ContainSubstring("recursive_merge!"))
							Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(rubyExceptionPrefixTemplate, erbTemplateFilepath)))
						})
					})
				})
			})
		})

		Describe("interactions with FileSystem", func() {
			var (
				fakeFs     *erbrendererfakes.FakeFileSystem
				fakeRunner *erbrendererfakes.FakeCmdRunner
			)

			BeforeEach(func() {
				fakeFs = &erbrendererfakes.FakeFileSystem{}
				fakeRunner = &erbrendererfakes.FakeCmdRunner{}

				erbTemplateContent = ""
				err := os.WriteFile(erbTemplateFilepath, []byte(erbTemplateContent), 0666)
				Expect(err).ToNot(HaveOccurred())
				erbRenderer = erbrenderer.NewERBRenderer(fakeFs, fakeRunner, logger)
			})

			It("cleans up temporary directory", func() {
				err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
				Expect(err).ToNot(HaveOccurred())
				Expect(fs.FileExists("fake-temp-dir")).To(BeFalse())
			})

			Context("when creating temporary directory fails", func() {
				var tempDirErr error

				BeforeEach(func() {
					tempDirErr = errors.New("fake-temp-dir-err")
					fakeFs.TempDirReturns("", tempDirErr)
				})

				It("returns an error", func() {
					err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(tempDirErr.Error()))
				})
			})

			Context("when writing renderer script fails", func() {
				var writerErr error

				BeforeEach(func() {
					writerErr = errors.New("fake-writer-err")
					fakeFs.WriteFileStringReturns(writerErr)
				})

				It("returns an error", func() {
					err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Writing renderer script: fake-writer-err"))
				})
			})

			Context("when writing renderer context fails", func() {
				var writerErr error

				BeforeEach(func() {
					writerErr = errors.New("fake-writer-err")
					fakeFs.WriteFileStringReturnsOnCall(0, nil)
					fakeFs.WriteFileStringReturnsOnCall(1, writerErr)
				})

				It("returns an error", func() {
					err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Writing context: fake-writer-err"))
				})
			})
		})

		Describe("interactions with CmdRunner", func() {
			var (
				fakeRunner *erbrendererfakes.FakeCmdRunner
			)

			BeforeEach(func() {
				fakeRunner = &erbrendererfakes.FakeCmdRunner{}
				erbRenderer = erbrenderer.NewERBRenderer(fs, fakeRunner, logger)
			})

			It("constructs ruby erb rendering command", func() {
				err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeRunner.RunComplexCommandCallCount()).To(Equal(1))
				command := fakeRunner.RunComplexCommandArgsForCall(0)
				Expect(command.Name).To(Equal("ruby"))
				Expect(command.Args[0]).To(MatchRegexp("/erb_renderer\\.rb$"))
				Expect(command.Args[1]).To(MatchRegexp("/erb-context\\.json$"))
				Expect(command.Args[2]).To(Equal(erbTemplateFilepath))
				Expect(command.Args[3]).To(Equal(renderedTemplatePath))
			})

			Context("when running ruby command fails", func() {
				BeforeEach(func() {
					fakeRunner.RunComplexCommandReturns("", "", 1, errors.New("fake-cmd-error"))
				})

				It("returns an error", func() {
					err := erbRenderer.Render(erbTemplateFilepath, renderedTemplatePath, context)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-cmd-error"))
				})
			})
		})
	})
})
