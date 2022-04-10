package storage

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestStorage(t *testing.T) {
	RegisterTestingT(t)

	type dataObject struct {
		Param string
	}

	{
		storage, err := New[dataObject](Config{
			FilePath: "./test",
		})
		Expect(err).ShouldNot(HaveOccurred())

		d := storage.Load()
		Expect(d.Param).To(Equal(""))
		d.Param = "value"

		err = storage.Save(d)
		Expect(err).ShouldNot(HaveOccurred())
	}

	{
		storage, err := New[dataObject](Config{
			FilePath: "./test",
		})
		Expect(err).ShouldNot(HaveOccurred())

		d := storage.Load()
		Expect(d.Param).To(Equal("value"))
	}

	err := os.Remove("./test")
	Expect(err).ShouldNot(HaveOccurred())

}
