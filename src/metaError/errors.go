package metaError

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
)

type apiError interface {
	GetMeta() interface{}
	GetMessage() string
}

// Print prints out meta in the os.Stderr
func Print(err error) {
	var apiErr apiError
	if !errors.As(err, &apiErr) {
		fmt.Fprintln(os.Stderr, "error:", err)
		return
	}
	fmt.Fprintln(os.Stderr, "error:", strings.ToLower(apiErr.GetMessage()))
	if apiErr.GetMeta() != nil {
		meta, err := yaml.Marshal(apiErr.GetMeta())
		if err != nil {
			fmt.Fprintln(os.Stderr, "couldn't parse meta")
		}
		fmt.Fprintln(os.Stderr, string(meta))
	}
}
