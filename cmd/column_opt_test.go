package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("ColumnOpt", func() {
	It("parses given header", func() {
		var columnOpt ColumnOpt
		columnOpt.UnmarshalFlag("Header1")
		Expect(columnOpt.Header).To(Equal(boshtbl.NewHeader("Header1")))
	})
})
