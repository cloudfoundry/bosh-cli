package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
)

var _ = Describe("URLArg", func() {
	Describe("IsEmpty", func() {
		It("returns true if empty", func() {
			Expect(URLArg("val").IsEmpty()).To(BeFalse())
			Expect(URLArg("").IsEmpty()).To(BeTrue())
		})
	})

	Describe("IsRemote", func() {
		It("returns true if http/https scheme is used", func() {
			Expect(URLArg("https://host").IsRemote()).To(BeTrue())
			Expect(URLArg("http://host").IsRemote()).To(BeTrue())
			Expect(URLArg("other://host").IsRemote()).To(BeFalse())
		})
	})

	Describe("FilePath", func() {
		It("returns path without 'file://'", func() {
			Expect(URLArg("path").FilePath()).To(Equal("path"))
			Expect(URLArg("file://path").FilePath()).To(Equal("path"))
		})
	})
})
