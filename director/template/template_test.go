package template_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/director/template"
)

var _ = Describe("Template", func() {
	It("can template values into a byte slice", func() {
		template := NewTemplate([]byte("((key))"))
		variables := Variables{
			"key": "foo",
		}

		result, err := template.Evaluate(variables)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("foo\n")))
	})

	It("can template multiple values into a byte slice", func() {
		template := NewTemplate([]byte("((key)): ((value))"))
		variables := Variables{
			"key":   "foo",
			"value": "bar",
		}

		result, err := template.Evaluate(variables)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("foo: bar\n")))
	})

	It("can template boolean values into a byte slice", func() {
		template := NewTemplate([]byte("otherstuff: ((boule))"))
		variables := Variables{
			"boule": true,
		}
		result, err := template.Evaluate(variables)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("otherstuff: true\n")))
	})

	It("can template a different data types into a byte slice", func() {
		hashValue := map[string]interface{}{"key2": []string{"value1", "value2"}}
		template := NewTemplate([]byte("name1: ((name1))\nname2: ((name2))\nname3: ((name3))\nname4: ((name4))\nname5: ((name5))\nname6: ((name6))\n1234: value\n"))
		variables := Variables{
			"name1": 1,
			"name2": "nil",
			"name3": true,
			"name4": "",
			"name5": nil,
			"name6": map[string]interface{}{"key": hashValue},
		}
		result, err := template.Evaluate(variables)
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

	Context("When template is a string", func() {
		It("returns it", func() {
			template := NewTemplate([]byte(`"string with a ((key))"`))
			variables := Variables{
				"key": "not key",
			}
			result, err := template.Evaluate(variables)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]byte("string with a ((key))\n")))
		})
	})
	Context("When template is a number", func() {
		It("returns it", func() {
			template := NewTemplate([]byte(`1234`))
			variables := Variables{
				"key": "not key",
			}
			result, err := template.Evaluate(variables)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]byte("1234\n")))
		})
	})

	Context("When variable has nil as value for key", func() {
		It("uses null", func() {
			template := NewTemplate([]byte("((key)): value"))
			variables := Variables{
				"key": nil,
			}

			result, err := template.Evaluate(variables)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]byte("null: value\n")))
		})
	})

	It("can template unicode values into a byte slice", func() {
		template := NewTemplate([]byte("((Ω))"))
		variables := Variables{
			"Ω": "☃",
		}

		result, err := template.Evaluate(variables)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("☃\n")))
	})

	It("can template keys with dashes and underscores into a byte slice", func() {
		template := NewTemplate([]byte("((with-a-dash)): ((with_an_underscore))"))
		variables := Variables{
			"with-a-dash":        "dash",
			"with_an_underscore": "underscore",
		}

		result, err := template.Evaluate(variables)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("dash: underscore\n")))
	})

	It("can template the same value multiple times into a byte slice", func() {
		template := NewTemplate([]byte("((key)): ((key))"))
		variables := Variables{
			"key": "foo",
		}

		result, err := template.Evaluate(variables)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("foo: foo\n")))
	})

	It("can template values with strange newlines", func() {
		template := NewTemplate([]byte("((key))"))
		variables := Variables{
			"key": "this\nhas\nmany\nlines",
		}

		result, err := template.Evaluate(variables)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("|-\n  this\n  has\n  many\n  lines\n")))
	})

	It("ignores an invalid input", func() {
		template := NewTemplate([]byte("(()"))
		variables := Variables{}

		result, err := template.Evaluate(variables)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]byte("(()\n")))
	})
})
