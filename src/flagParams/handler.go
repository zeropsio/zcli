package flagParams

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zeropsio/zcli/src/constants"
	"github.com/zeropsio/zcli/src/gn"
)

type ParamsReader interface {
	IsSet(key string) bool
	HasSet(keys ...string) bool
	AllSet(keys ...string) bool

	GetString(name string) string
	GetStringSlice(name string) []string
	GetInt(name string) int
	GetBool(name string) bool
	GetLocalZCliYamlFileName() (string, bool)
}

var _ ParamsReader = (*Handler)(nil)

type Handler struct {
	viper                 *viper.Viper
	localZCliYamlFileName string
}

func New() *Handler {
	v := viper.New()
	v.SetEnvPrefix("ZEROPS")
	v.AutomaticEnv()

	cliDataPath, _, err := constants.ZcliYamlFilePath()
	if err == nil {
		v.SetConfigFile(cliDataPath)
	}
	if err := v.ReadInConfig(); err == nil {
		// Only print config file info for non-machine-readable output
		if v.GetString("format") != "json" {
			fmt.Fprintln(os.Stderr, "Using config file:", v.ConfigFileUsed()) // TODO (lh): log instead of print to stderr
		}
	}

	h := &Handler{
		viper: v,
	}

	v.AddConfigPath(".")
	v.SetConfigName(constants.CliZcliYamlBaseFileName)
	v.SetConfigType("yaml")
	if err := v.MergeInConfig(); err == nil {
		// Only print config file info for non-machine-readable output
		if v.GetString("format") != "json" {
			fmt.Fprintln(os.Stderr, "Using config file:", v.ConfigFileUsed()) // TODO (lh): log instead of print to stderr
		}
		h.localZCliYamlFileName = v.ConfigFileUsed()
	}

	return h
}

func (h *Handler) GetLocalZCliYamlFileName() (string, bool) {
	return h.localZCliYamlFileName, h.localZCliYamlFileName != ""
}

func (h *Handler) Bind(cmd *cobra.Command) {
	_ = h.viper.BindPFlags(cmd.Flags())
}

func (h *Handler) HasSet(flags ...string) bool {
	for _, f := range flags {
		if h.viper.IsSet(f) {
			return true
		}
	}
	return false
}

func (h *Handler) AllSet(flags ...string) bool {
	for _, f := range flags {
		if !h.viper.IsSet(f) {
			return false
		}
	}
	return true
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// kebabCaseToCamelCase this function here for backward compatibility for
// retrieving camel case values from envs or configs
func kebabCaseToCamelCase(kebab string) string {
	split := strings.Split(kebab, "-")
	camel := strings.ToLower(split[0])
	for _, s := range split[1:] {
		if s == "" {
			continue
		}
		camel += capitalize(s)
	}
	return camel
}

func (h *Handler) IsSet(key string) bool {
	return h.viper.IsSet(key) || h.viper.IsSet(kebabCaseToCamelCase(key))
}

func (h *Handler) GetString(name string) string {
	value := h.viper.GetString(kebabCaseToCamelCase(name))
	if !gn.IsZero(value) {
		return value
	}
	return h.viper.GetString(name)
}

func (h *Handler) GetStringSlice(name string) []string {
	value := h.viper.GetStringSlice(kebabCaseToCamelCase(name))
	if len(value) > 0 {
		return value
	}
	return h.viper.GetStringSlice(name)
}

func (h *Handler) GetInt(name string) int {
	value := h.viper.GetInt(kebabCaseToCamelCase(name))
	if !gn.IsZero(value) {
		return value
	}
	return h.viper.GetInt(name)
}

func (h *Handler) GetBool(name string) bool {
	value := h.viper.GetBool(kebabCaseToCamelCase(name))
	if !gn.IsZero(value) {
		return value
	}
	return h.viper.GetBool(name)
}
