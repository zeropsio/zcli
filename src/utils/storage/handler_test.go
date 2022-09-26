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

		{
			d := storage.Data()
			Expect(d.Param).To(Equal(""))
		}

		{
			d, err := storage.Update(func(data dataObject) dataObject {
				data.Param = "value"
				return data
			})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(d.Param).To(Equal("value"))
		}
	}

	{
		storage, err := New[dataObject](Config{
			FilePath: "./test",
		})
		Expect(err).ShouldNot(HaveOccurred())

		d := storage.Data()
		Expect(d.Param).To(Equal("value"))
	}

	err := os.Remove("./test")
	Expect(err).ShouldNot(HaveOccurred())

}
