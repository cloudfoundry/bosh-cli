package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakecmd "github.com/cloudfoundry/bosh-cli/v7/cmd/cmdfakes"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
)

var _ = Describe("LogInCmd", func() {
	var (
		basic    *fakecmd.FakeLoginStrategy
		uaa      *fakecmd.FakeLoginStrategy
		director *fakedir.FakeDirector
		command  cmd.LogInCmd
	)

	BeforeEach(func() {
		basic = &fakecmd.FakeLoginStrategy{}
		uaa = &fakecmd.FakeLoginStrategy{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewLogInCmd(basic, uaa, director)
	})

	Describe("Run", func() {
		act := func() error { return command.Run() }

		Context("when director uses basic auth", func() {
			BeforeEach(func() {
				director.InfoReturns(boshdir.Info{
					Auth: boshdir.UserAuthentication{Type: "basic"},
				}, nil)
			})

			It("uses basic login strategy", func() {
				basic.TryReturns(errors.New("fake-err"))
				Expect(act()).To(Equal(errors.New("fake-err")))
			})
		})

		Context("when director uses uaa auth", func() {
			BeforeEach(func() {
				director.InfoReturns(boshdir.Info{
					Auth: boshdir.UserAuthentication{Type: "uaa"},
				}, nil)
			})

			It("uses uaa login strategy", func() {
				uaa.TryReturns(errors.New("fake-err"))
				Expect(act()).To(Equal(errors.New("fake-err")))
			})
		})

		Context("when director uses unknown auth", func() {
			BeforeEach(func() {
				director.InfoReturns(boshdir.Info{
					Auth: boshdir.UserAuthentication{Type: "other"},
				}, nil)
			})

			It("returns an error", func() {
				Expect(act()).To(Equal(errors.New("Unknown auth type 'other'")))
			})
		})
	})
})
