package flagParams

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zeropsio/zcli/src/constants"
)

type Handler struct {
	*viper.Viper
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
		fmt.Fprintln(os.Stderr, "Using config file:", v.ConfigFileUsed()) // TODO (lh): log instead of print to stderr
	}

	v.AddConfigPath(".")
	v.SetConfigName(constants.CliZcliYamlBaseFileName)
	v.SetConfigType("yaml")
	if err := v.MergeInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", v.ConfigFileUsed()) // TODO (lh): log instead of print to stderr
	}

	return &Handler{
		Viper: v,
	}
}

func (h *Handler) Bind(cmd *cobra.Command) {
	_ = h.BindPFlags(cmd.Flags())
}

func (h *Handler) HasSet(flags ...string) bool {
	for _, f := range flags {
		if h.IsSet(f) {
			return true
		}
	}
	return false
}

func (h *Handler) AllSet(flags ...string) bool {
	for _, f := range flags {
		if !h.IsSet(f) {
			return false
		}
	}
	return true
}
