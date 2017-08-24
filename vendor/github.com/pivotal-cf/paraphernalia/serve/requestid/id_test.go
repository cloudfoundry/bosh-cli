package requestid_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/paraphernalia/serve/requestid"
)

var _ = Describe("generation", func() {
	It("generates unique IDs", func() {
		ids := make(map[string]struct{})

		for i := 0; i < 1000; i++ {
			id := requestid.Generate()
			ids[id] = struct{}{}
		}

		Expect(ids).To(HaveLen(1000))
	})
})

func BenchmarkGenerate(b *testing.B) {
	var i string

	for n := 0; n < b.N; n++ {
		// store result into variable so functionc all cannot be optimized away
		i = requestid.Generate()
	}

	// store result in package variable so the benchmark cannot be optimized away
	id = i
}

var id string
