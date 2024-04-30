package upgrade

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCompareVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compare version suite")
}

var _ = Describe("Upgrade", func() {
	Describe("compareVersion", func() {
		Context("when comparing version", func() {
			It("should return true when image tag is newer", func() {
				osLines := []string{"IMAGE=t1", "IMAGE_TAG=v1.0.1"}
				hostLines := []string{"IMAGE=t1", "IMAGE_TAG=v1.0.0"}
				result, err := compareVersion(osLines, hostLines)
				Expect(result).To(BeTrue())
				Expect(err).To(BeNil())
			})
			It("should return true when lines are not in order", func() {
				osLines := []string{"IMAGE_TAG=v1.0.1", "IMAGE=t1"}
				hostLines := []string{"IMAGE=t1", "IMAGE_TAG=v1.0.0"}
				result, err := compareVersion(osLines, hostLines)
				Expect(result).To(BeTrue())
				Expect(err).To(BeNil())
			})

			It("should return false when image is not the same", func() {
				osLines := []string{"IMAGE=t1.1", "IMAGE_TAG=v1.0.1"}
				hostLines := []string{"IMAGE=t1.2", "IMAGE_TAG=v1.0.0"}
				result, err := compareVersion(osLines, hostLines)
				Expect(result).To(BeFalse())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("OS image is different"))
			})

			It("should return false when container image tag is older", func() {
				osLines := []string{"IMAGE=t2", "IMAGE_TAG=v1.0.0"}
				hostLines := []string{"IMAGE=t2", "IMAGE_TAG=v1.0.1"}
				result, err := compareVersion(osLines, hostLines)
				Expect(result).To(BeFalse())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("current OS image tag is either older or equal to the host OS image tag"))
			})

			It("should return false when image tag is equal", func() {
				osLines := []string{"IMAGE=t3", "IMAGE_TAG=v1.0.0"}
				hostLines := []string{"IMAGE=t3", "IMAGE_TAG=v1.0.0"}
				result, err := compareVersion(osLines, hostLines)
				Expect(result).To(BeFalse())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("current OS image tag is either older or equal to the host OS image tag"))
			})

			It("should return error when osLines is empty", func() {
				osLines := []string{}
				hostLines := []string{"IMAGE=t4", "IMAGE_TAG=v1.0.0"}
				result, err := compareVersion(osLines, hostLines)
				Expect(result).To(BeFalse())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("either image or host OS release file is empty"))
			})

			It("should return error when hostLines is empty", func() {
				osLines := []string{"IMAGE=t5", "IMAGE_TAG=v1.0.0"}
				hostLines := []string{}
				result, err := compareVersion(osLines, hostLines)
				Expect(result).To(BeFalse())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("either image or host OS release file is empty"))
			})

			It("should return error when both osLines and hostLines are empty", func() {
				osLines := []string{}
				hostLines := []string{}
				result, err := compareVersion(osLines, hostLines)
				Expect(result).To(BeFalse())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("either image or host OS release file is empty"))
			})
			It("should return false when only IMAGE_TAG is provided", func() {
				osLines := []string{"IMAGE_TAG=v1.0.1"}
				hostLines := []string{"IMAGE_TAG=v1.0.0"}
				result, err := compareVersion(osLines, hostLines)
				Expect(result).To(BeFalse())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to compare OS release"))
			})
			It("should return false when only IMAGE_TAG is invalid", func() {
				osLines := []string{"IMAGE=t6", "IMAGE_TAG=abc"}
				hostLines := []string{"IMAGE=t6", "IMAGE_TAG=acd"}
				result, err := compareVersion(osLines, hostLines)
				Expect(result).To(BeFalse())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse version"))
			})
		})
	})
})
