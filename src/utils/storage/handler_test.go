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
		storage, err := New(Config{
			FilePath: "./test",
		})
		Expect(err).ShouldNot(HaveOccurred())

		data := storage.Load(&dataObject{})
		if d, ok := data.(*dataObject); ok {
			Expect(d.Param).To(Equal(""))
			d.Param = "value"

			err = storage.Save(d)
			Expect(err).ShouldNot(HaveOccurred())
		} else {
			t.Fail()
		}
	}

	{
		storage, err := New(Config{
			FilePath: "./test",
		})
		Expect(err).ShouldNot(HaveOccurred())

		data := storage.Load(&dataObject{})
		if d, ok := data.(*dataObject); ok {
			Expect(d.Param).To(Equal("value"))
		} else {
			t.Fail()
		}
	}

	err := os.Remove("./test")
	Expect(err).ShouldNot(HaveOccurred())

}
