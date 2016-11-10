package template_test

import (
	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/director/template"
)

var _ = Describe("Template", func() {
	It("can interpolate values into a struct with byte slice", func() {
		template := NewTemplate([]byte("((key))"))
		vars := Variables{"key": "foo"}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("foo\n")))
	})

	It("can interpolate multiple values into a byte slice", func() {
		template := NewTemplate([]byte("((key)): ((value))"))
		vars := Variables{
			"key":   "foo",
			"value": "bar",
		}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("foo: bar\n")))
	})

	It("can interpolate boolean values into a byte slice", func() {
		template := NewTemplate([]byte("otherstuff: ((boule))"))
		vars := Variables{"boule": true}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("otherstuff: true\n")))
	})

	It("can interpolate a different data types into a byte slice", func() {
		hashValue := map[string]interface{}{"key2": []string{"value1", "value2"}}
		template := NewTemplate([]byte("name1: ((name1))\nname2: ((name2))\nname3: ((name3))\nname4: ((name4))\nname5: ((name5))\nname6: ((name6))\n1234: value\n"))
		vars := Variables{
			"name1": 1,
			"name2": "nil",
			"name3": true,
			"name4": "",
			"name5": nil,
			"name6": map[string]interface{}{"key": hashValue},
		}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte(`1234: value
name1: 1
name2: nil
name3: true
name4: ""
name5: null
name6:
  key:
    key2:
    - value1
    - value2
`)))
	})

	It("can interpolate different data types into a byte slice with !key", func() {
		template := NewTemplate([]byte("otherstuff: ((!boule))"))
		vars := Variables{"boule": true}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("otherstuff: true\n")))
	})

	It("return errors if there are missing variable keys and ExpectAllKeys is true", func() {
		template := NewTemplate([]byte(`
((key)): ((key2))
((key3)): 2
dup-key: ((key3))
((key4))_array:
- ((key_in_array))
`))
		vars := Variables{"key3": "foo"}

		_, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{ExpectAllKeys: true})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Expected to find variables: key, key2, key4, key_in_array"))
	})

	It("does not return error if there are missing variable keys and ExpectAllKeys is false", func() {
		template := NewTemplate([]byte("((key)): ((key2))\n((key3)): 2"))
		vars := Variables{"key3": "foo"}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal([]byte("((key)): ((key2))\nfoo: 2\n")))
	})

	Context("When template is a number", func() {
		It("returns it", func() {
			template := NewTemplate([]byte(`1234`))
			vars := Variables{"key": "not key"}

			result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]byte("1234\n")))
		})
	})

	Context("When variable has nil as value for key", func() {
		It("uses null", func() {
			template := NewTemplate([]byte("((key)): value"))
			vars := Variables{"key": nil}

			result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]byte("null: value\n")))
		})
	})

	It("can interpolate unicode values into a byte slice", func() {
		template := NewTemplate([]byte("((Ω))"))
		vars := Variables{"Ω": "☃"}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("☃\n")))
	})

	It("can interpolate keys with dashes and underscores into a byte slice", func() {
		template := NewTemplate([]byte("((with-a-dash)): ((with_an_underscore))"))
		vars := Variables{
			"with-a-dash":        "dash",
			"with_an_underscore": "underscore",
		}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("dash: underscore\n")))
	})

	It("can interpolate a secret key in the middle of a string", func() {
		template := NewTemplate([]byte("url: https://((ip))"))
		vars := Variables{
			"ip": "10.0.0.0",
		}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("url: https://10.0.0.0\n")))
	})

	It("can interpolate multiple secret keys in the middle of a string", func() {
		template := NewTemplate([]byte("uri: nats://nats:((password))@((ip)):4222"))
		vars := Variables{
			"password": "secret",
			"ip":       "10.0.0.0",
		}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("uri: nats://nats:secret@10.0.0.0:4222\n")))
	})

	It("can interpolate multiple keys of type string and int in the middle of a string", func() {
		template := NewTemplate([]byte("address: ((ip)):((port))"))
		vars := Variables{
			"port": 4222,
			"ip":   "10.0.0.0",
		}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("address: 10.0.0.0:4222\n")))
	})

	It("raises error when interpolating an unsupported type in the middle of a string", func() {
		template := NewTemplate([]byte("address: ((definition)):((eulers_number))"))
		vars := Variables{
			"eulers_number": 2.717,
			"definition":    "natural_log",
		}

		_, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("2.717"))
		Expect(err.Error()).To(ContainSubstring("eulers_number"))
	})

	It("can interpolate a single key multiple times in the middle of a string", func() {
		template := NewTemplate([]byte("acct_and_password: ((user)):((user))"))
		vars := Variables{
			"user": "nats",
		}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("acct_and_password: nats:nats\n")))
	})

	It("can interpolate values into the middle of a key", func() {
		template := NewTemplate([]byte("((iaas))_cpi: props"))
		vars := Variables{
			"iaas": "aws",
		}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("aws_cpi: props\n")))
	})

	It("can interpolate the same value multiple times into a byte slice", func() {
		template := NewTemplate([]byte("((key)): ((key))"))
		vars := Variables{"key": "foo"}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("foo: foo\n")))
	})

	It("can interpolate values with strange newlines", func() {
		template := NewTemplate([]byte("((key))"))
		vars := Variables{"key": "this\nhas\nmany\nlines"}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("|-\n  this\n  has\n  many\n  lines\n")))
	})

	It("ignores an invalid input", func() {
		template := NewTemplate([]byte("(()"))
		vars := Variables{}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("(()\n")))
	})

	It("strips away ! from variable keys", func() {
		template := NewTemplate([]byte("abc: ((!key))\nxyz: [((!key))]"))
		vars := Variables{"key": "val"}

		result, err := template.Evaluate(vars, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("abc: val\nxyz:\n- val\n")))
	})

	It("can run operations to modify document", func() {
		template := NewTemplate([]byte("a: b"))
		vars := Variables{}
		ops := patch.Ops{
			patch.ReplaceOp{Path: patch.MustNewPointerFromString("/a"), Value: "c"},
		}

		result, err := template.Evaluate(vars, ops, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("a: c\n")))
	})

	It("interpolates after running operations", func() {
		template := NewTemplate([]byte("a: b"))
		vars := Variables{"c": "x"}
		ops := patch.Ops{
			patch.ReplaceOp{Path: patch.MustNewPointerFromString("/a"), Value: "((c))"},
		}

		result, err := template.Evaluate(vars, ops, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("a: x\n")))
	})

	It("returns an error if variables added by operations are not found", func() {
		template := NewTemplate([]byte("a: b"))
		vars := Variables{}
		ops := patch.Ops{
			patch.ReplaceOp{Path: patch.MustNewPointerFromString("/a"), Value: "((c))"},
		}

		_, err := template.Evaluate(vars, ops, EvaluateOpts{ExpectAllKeys: true})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Expected to find variables: c"))
	})

	It("returns an error if operation fails", func() {
		template := NewTemplate([]byte("a: b"))
		vars := Variables{}
		ops := patch.Ops{
			patch.ReplaceOp{Path: patch.MustNewPointerFromString("/x/y"), Value: "c"},
		}

		_, err := template.Evaluate(vars, ops, EvaluateOpts{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Expected to find a map key 'x' for path '/x'"))
	})

	It("returns raw bytes of a string if UnescapedMultiline is true", func() {
		template := NewTemplate([]byte("value"))

		result, err := template.Evaluate(Variables{}, patch.Ops{}, EvaluateOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("value\n")))

		result, err = template.Evaluate(Variables{}, patch.Ops{}, EvaluateOpts{UnescapedMultiline: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("value\n")))
	})
})
