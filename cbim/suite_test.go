package cbim

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"V1Beta1 Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = Describe("GlobalSecondaryIndexIdentifier.ToString", func() {

	It("should exclude the default collection name", func() {
		// Arrange

		identifier := GlobalSecondaryIndexIdentifier{
			ScopeName:      "_default",
			CollectionName: "_default",
			Name:           "my_index",
		}

		// Act

		result := identifier.ToString()

		// Assert

		Expect(result).To(Equal("my_index"))
	})

	It("should return a dotted name", func() {
		// Arrange

		identifier := GlobalSecondaryIndexIdentifier{
			ScopeName:      "scope",
			CollectionName: "_default",
			Name:           "my_index",
		}

		// Act

		result := identifier.ToString()

		// Assert

		Expect(result).To(Equal("scope._default.my_index"))
	})
})

var _ = Describe("ParseIndexIdentifierString", func() {

	It("should error on empty string", func() {
		// Act

		_, err := ParseIndexIdentifierString("")

		// Assert

		Expect(err).NotTo(BeNil())
	})

	It("should error on empty segment", func() {
		// Act

		_, err := ParseIndexIdentifierString("scope..name")

		// Assert

		Expect(err).NotTo(BeNil())
	})

	It("should error on two segments", func() {
		// Act

		_, err := ParseIndexIdentifierString("scope.name")

		// Assert

		Expect(err).NotTo(BeNil())
	})

	It("should error on four segments", func() {
		// Act

		_, err := ParseIndexIdentifierString("scope.collection.name.extra")

		// Assert

		Expect(err).NotTo(BeNil())
	})

	It("should return simple name in default collection", func() {
		// Act

		result, err := ParseIndexIdentifierString("name")

		// Assert

		Expect(result).To(Equal(GlobalSecondaryIndexIdentifier{
			ScopeName:      "_default",
			CollectionName: "_default",
			Name:           "name",
		}))
		Expect(err).To(BeNil())
	})

	It("should return dotted name", func() {
		// Act

		result, err := ParseIndexIdentifierString("scope.collection.name")

		// Assert

		Expect(result).To(Equal(GlobalSecondaryIndexIdentifier{
			ScopeName:      "scope",
			CollectionName: "collection",
			Name:           "name",
		}))
		Expect(err).To(BeNil())
	})
})
