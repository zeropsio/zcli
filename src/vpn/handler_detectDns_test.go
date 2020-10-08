package vpn

import (
	"testing"

	. "github.com/onsi/gomega"
)

var testErrorResponseDataProvider = []struct {
	name     string
	filePath string
	expected bool
}{
	{name: "valid", filePath: "./detectDnsTestFiles/test1", expected: true},
	{name: "no nameserver 127.0.0.53", filePath: "./detectDnsTestFiles/test2", expected: false},
	{name: "wrong nameserver order", filePath: "./detectDnsTestFiles/test3", expected: false},
}

func TestIsValidSystemdResolve(t *testing.T) {
	for _, test := range testErrorResponseDataProvider {
		test := test // scope lint
		t.Run(test.name, func(t *testing.T) {
			RegisterTestingT(t)

			result, err := isValidSystemdResolve(test.filePath)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(result).To(Equal(test.expected))
		})
	}
}
