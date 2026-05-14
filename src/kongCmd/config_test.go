package kongCmd

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/require"

	"github.com/zeropsio/zcli/src/constants"
)

// loadInDir constructs a Config with:
//   - localDir as the CWD (where the local zcli.yaml is read from)
//   - globalPath as the value of CliZcliYamlFilePathEnvVar (so the
//     constants package returns it from ZcliYamlFilePath)
//
// Both are temp paths managed by the test runtime.
func loadInDir(t *testing.T, localDir, globalPath string) *Config {
	t.Helper()
	t.Chdir(localDir)
	t.Setenv(constants.CliZcliYamlFilePathEnvVar, globalPath)
	return LoadConfig(io.Discard)
}

// resolveStringFlag asks the Config to resolve a fake string flag named `name`.
// Returns the raw resolver result (nil means "no source has this flag").
func resolveStringFlag(t *testing.T, c *Config, name string) any {
	t.Helper()
	v, err := c.Resolve(nil, nil, &kong.Flag{Value: &kong.Value{Name: name}})
	require.NoError(t, err)
	return v
}

func TestConfig_NoFilesIsEmpty(t *testing.T) {
	c := loadInDir(t, t.TempDir(), filepath.Join(t.TempDir(), "missing.yml"))
	require.Nil(t, resolveStringFlag(t, c, "anything"))
	require.False(t, c.IsSet("anything"))
	_, ok := c.LocalZcliYamlFile()
	require.False(t, ok)
}

func TestConfig_GlobalOnly(t *testing.T) {
	global := filepath.Join(t.TempDir(), "zcli.yml")
	require.NoError(t, os.WriteFile(global, []byte("working-dir: /global\n"), 0o600))

	c := loadInDir(t, t.TempDir(), global)

	require.Equal(t, "/global", resolveStringFlag(t, c, "working-dir"))
	require.True(t, c.IsSet("working-dir"))
}

func TestConfig_LocalOverridesGlobal(t *testing.T) {
	globalDir := t.TempDir()
	global := filepath.Join(globalDir, "zcli.yml")
	require.NoError(t, os.WriteFile(global, []byte("working-dir: /global\nsetup: from-global\n"), 0o600))

	local := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(local, ".zcli.yml"), []byte("working-dir: /local\n"), 0o600))

	c := loadInDir(t, local, global)

	require.Equal(t, "/local", resolveStringFlag(t, c, "working-dir"),
		"local must override global on key collision")
	require.Equal(t, "from-global", resolveStringFlag(t, c, "setup"),
		"keys only in global are still resolved")
	gotPath, ok := c.LocalZcliYamlFile()
	require.True(t, ok)
	require.Equal(t, filepath.Join(".", ".zcli.yml"), gotPath)
}

func TestConfig_LocalYamlExtensionFallback(t *testing.T) {
	local := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(local, ".zcli.yaml"),
		[]byte("setup: yaml-ext\n"), 0o600))

	c := loadInDir(t, local, "")

	require.Equal(t, "yaml-ext", resolveStringFlag(t, c, "setup"))
}

func TestConfig_EnvOverridesYaml(t *testing.T) {
	local := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(local, ".zcli.yml"),
		[]byte("working-dir: /from-yaml\n"), 0o600))

	t.Setenv("ZEROPS_WORKING_DIR", "/from-env")

	c := loadInDir(t, local, "")

	require.Equal(t, "/from-env", resolveStringFlag(t, c, "working-dir"),
		"env must win over yaml")
	require.True(t, c.IsSet("working-dir"))
}

// Regression for the cli-tests bugfix in createRunFunc: an exported-but-empty
// ZEROPS_TOKEN must not silently log the user out. Same semantics applied to
// all ZEROPS_* env vars in our resolver.
func TestConfig_EmptyEnvIgnored(t *testing.T) {
	local := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(local, ".zcli.yml"),
		[]byte("working-dir: /from-yaml\n"), 0o600))

	t.Setenv("ZEROPS_WORKING_DIR", "")

	c := loadInDir(t, local, "")

	require.Equal(t, "/from-yaml", resolveStringFlag(t, c, "working-dir"),
		"empty env must not shadow yaml")
}

func TestConfig_CamelCaseFallback(t *testing.T) {
	local := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(local, ".zcli.yml"),
		[]byte("workingDir: /from-camel\n"), 0o600))

	c := loadInDir(t, local, "")

	require.Equal(t, "/from-camel", resolveStringFlag(t, c, "working-dir"),
		"kebab-named flag must find a camelCase yaml key as fallback")
	require.True(t, c.IsSet("working-dir"))
}

func TestConfig_KebabKeyWinsOverCamel(t *testing.T) {
	local := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(local, ".zcli.yml"),
		[]byte("working-dir: /kebab\nworkingDir: /camel\n"), 0o600))

	c := loadInDir(t, local, "")

	require.Equal(t, "/kebab", resolveStringFlag(t, c, "working-dir"),
		"kebab-form yaml key must win when both forms are present")
}

func TestConfig_StringSliceFromYaml(t *testing.T) {
	local := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(local, ".zcli.yml"),
		[]byte("path-to-file-or-dir:\n  - a\n  - b\n  - c\n"), 0o600))

	c := loadInDir(t, local, "")

	got := resolveStringFlag(t, c, "path-to-file-or-dir")
	require.Equal(t, []string{"a", "b", "c"}, got,
		"yaml []any of strings must be normalized to []string for kong")
}

func TestConfig_MixedSliceLeftAsIs(t *testing.T) {
	v := normalizeYamlValue([]any{"a", 1, true})
	require.Equal(t, []any{"a", 1, true}, v,
		"slices with mixed types should pass through unchanged")
}

func TestConfig_IsSetForCamelCase(t *testing.T) {
	local := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(local, ".zcli.yml"),
		[]byte("workingDir: /camel\n"), 0o600))

	c := loadInDir(t, local, "")

	require.True(t, c.IsSet("working-dir"),
		"IsSet must report true for the kebab name even when only camelCase is in yaml")
}

func TestUpperSnakeAndKebabToCamel(t *testing.T) {
	cases := []struct {
		kebab     string
		camel     string
		envSuffix string
	}{
		{"working-dir", "workingDir", "WORKING_DIR"},
		{"zerops-yaml-path", "zeropsYamlPath", "ZEROPS_YAML_PATH"},
		{"setup", "setup", "SETUP"},
		{"path-to-file-or-dir", "pathToFileOrDir", "PATH_TO_FILE_OR_DIR"},
	}
	for _, tc := range cases {
		t.Run(tc.kebab, func(t *testing.T) {
			if got := kebabToCamel(tc.kebab); got != tc.camel {
				t.Fatalf("kebabToCamel(%q) = %q, want %q", tc.kebab, got, tc.camel)
			}
			if got := upperSnake(tc.kebab); got != tc.envSuffix {
				t.Fatalf("upperSnake(%q) = %q, want %q", tc.kebab, got, tc.envSuffix)
			}
		})
	}
}

// Compile-time check: Config must implement kong.Resolver as imported.
var _ kong.Resolver = (*Config)(nil)

// Compile-time check: normalizeYamlValue returns interface{}, but the [...]string
// path goes through reflect for the table-driven test above. Keep the helper
// honest by re-checking the typed assertion once.
func TestNormalizeYamlValue_TypedReturn(t *testing.T) {
	got := normalizeYamlValue([]any{"x", "y"})
	if reflect.TypeOf(got).String() != "[]string" {
		t.Fatalf("expected []string, got %T", got)
	}
}
