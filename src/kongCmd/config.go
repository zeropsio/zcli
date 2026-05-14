package kongCmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"gopkg.in/yaml.v3"

	"github.com/zeropsio/zcli/src/constants"
)

// Config replaces the old viper-backed flagParams. It loads a global yaml
// config and merges a local zcli.yaml on top, exposes them via a kong.Resolver,
// and answers IsSet for the few callers that need to distinguish user-set
// from default.
//
// Precedence (highest first): CLI > env (ZEROPS_*) > local yaml > global yaml
// > flag default. Kong consults the resolver only for unset CLI flags, so
// CLI wins automatically.
type Config struct {
	merged    map[string]any
	localFile string
}

// LoadConfig reads global + local yaml. notice receives one line per file
// read (pass os.Stderr for production, io.Discard in tests). The function
// never errors — missing files are normal.
func LoadConfig(notice io.Writer) *Config {
	if notice == nil {
		notice = io.Discard
	}
	c := &Config{merged: map[string]any{}}

	if globalPath, _, err := constants.ZcliYamlFilePath(); err == nil {
		if m, ok := readYaml(globalPath); ok {
			fmt.Fprintln(notice, "Using config file:", globalPath)
			for k, v := range m {
				c.merged[k] = v
			}
		}
	}

	base := constants.CliZcliYamlBaseFileName
	for _, ext := range []string{".yml", ".yaml"} {
		path := filepath.Join(".", base+ext)
		if m, ok := readYaml(path); ok {
			fmt.Fprintln(notice, "Using config file:", path)
			for k, v := range m {
				c.merged[k] = v
			}
			c.localFile = path
			break
		}
	}

	return c
}

func readYaml(path string) (map[string]any, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, false
	}
	if m == nil {
		return map[string]any{}, true
	}
	return m, true
}

// LocalZcliYamlFile returns the path to the local config file and whether
// one was found. Replaces flagParams.Handler.GetLocalZCliYamlFileName.
func (c *Config) LocalZcliYamlFile() (string, bool) {
	return c.localFile, c.localFile != ""
}

// Validate is part of the kong.Resolver interface. We have no resolver-wide
// constraints, so this is a no-op.
func (c *Config) Validate(_ *kong.Application) error { return nil }

// Resolve implements kong.Resolver. Called once per flag that has no CLI
// value. Tries env first (matching ZEROPS_<UPPER_SNAKE>), then yaml — kebab
// key first, then camelCase fallback for back-compat with older configs.
func (c *Config) Resolve(_ *kong.Context, _ *kong.Path, flag *kong.Flag) (any, error) {
	name := flag.Name

	if v := os.Getenv("ZEROPS_" + upperSnake(name)); v != "" {
		return v, nil
	}

	if v, ok := c.merged[name]; ok {
		return normalizeYamlValue(v), nil
	}
	if v, ok := c.merged[kebabToCamel(name)]; ok {
		return normalizeYamlValue(v), nil
	}
	return nil, nil
}

// IsSet reports whether name has a value from any source other than CLI
// (env or yaml). Used at the few sites that distinguish user-set from
// default — see servicePush and serviceDeploy.
func (c *Config) IsSet(name string) bool {
	if v := os.Getenv("ZEROPS_" + upperSnake(name)); v != "" {
		return true
	}
	if _, ok := c.merged[name]; ok {
		return true
	}
	if _, ok := c.merged[kebabToCamel(name)]; ok {
		return true
	}
	return false
}

// normalizeYamlValue converts []any (the default for yaml sequences) into
// []string when all elements are strings. Kong's slice mapper handles []string
// directly; mixed-type slices are passed through and may fail to decode.
func normalizeYamlValue(v any) any {
	arr, ok := v.([]any)
	if !ok {
		return v
	}
	strs := make([]string, 0, len(arr))
	for _, e := range arr {
		s, ok := e.(string)
		if !ok {
			return v
		}
		strs = append(strs, s)
	}
	return strs
}

func upperSnake(kebab string) string {
	return strings.ToUpper(strings.ReplaceAll(kebab, "-", "_"))
}

func kebabToCamel(kebab string) string {
	parts := strings.Split(kebab, "-")
	if len(parts) == 0 {
		return kebab
	}
	out := strings.ToLower(parts[0])
	for _, p := range parts[1:] {
		if p == "" {
			continue
		}
		out += strings.ToUpper(p[:1]) + p[1:]
	}
	return out
}
