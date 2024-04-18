package cmd_test

import (
	"github.com/llmos-ai/llmos/pkg/utils/cmd"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runner", Label("types", "runner"), func() {
	It("Runs commands on the real Runner", func() {
		r := cmd.NewRunner()
		_, err := r.Run("pwd")
		Expect(err).To(BeNil())
	})
	It("returns false if command does not exists", func() {
		r := cmd.NewRunner()
		exists := r.CmdExist("ABCDE")
		Expect(exists).To(BeFalse())
	})
	It("returns true if command exists", func() {
		r := cmd.NewRunner()
		exists := r.CmdExist("true")
		Expect(exists).To(BeTrue())
	})
})
