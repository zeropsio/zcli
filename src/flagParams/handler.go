package flagParams

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zeropsio/zcli/src/constants"
)

type Handler struct {
	params map[string]interface{}
	viper  *viper.Viper
}

func New() *Handler {
	return &Handler{
		params: make(map[string]interface{}),
		viper:  viper.New(),
	}
}

func (h *Handler) getCmdId(cmd *cobra.Command, name string) string {
	return cmd.Use + name
}

func (h *Handler) RegisterString(cmd *cobra.Command, name, shorthand, defaultValue, description string) {
	var paramValue string

	cmd.Flags().StringVarP(&paramValue, name, shorthand, defaultValue, description)

	h.params[h.getCmdId(cmd, name)] = func() *string {
		if cmd.Flags().Lookup(name).Changed {
			return &paramValue
		}
		if val := h.viper.GetString(toSnakeCase(name)); val != "" {
			return &val
		}
		return &paramValue
	}
}

func (h *Handler) RegisterBool(cmd *cobra.Command, name, shorthand string, defaultValue bool, description string) {
	var paramValue bool

	cmd.Flags().BoolVarP(&paramValue, name, shorthand, defaultValue, description)

	h.params[h.getCmdId(cmd, name)] = func() *bool {
		if cmd.Flags().Lookup(name).Changed {
			return &paramValue
		}
		if val := h.viper.GetBool(toSnakeCase(name)); val {
			return &val
		}
		return &paramValue
	}
}

func (h *Handler) RegisterInt(cmd *cobra.Command, name, shorthand string, defaultValue int, description string) {
	var paramValue int

	cmd.Flags().IntVarP(&paramValue, name, shorthand, defaultValue, description)

	h.params[h.getCmdId(cmd, name)] = func() *int {
		if cmd.Flags().Lookup(name).Changed {
			return &paramValue
		}
		if val := h.viper.GetInt(toSnakeCase(name)); val != 0 {
			return &val
		}
		return &paramValue
	}
}

func (h *Handler) GetString(cmd *cobra.Command, name string) string {
	id := h.getCmdId(cmd, name)
	if param, exists := h.params[id]; exists {
		if v, ok := param.(func() *string); ok {
			return *v()
		}
		return ""
	}
	return ""
}

func (h *Handler) GetInt(cmd *cobra.Command, name string) int {
	id := h.getCmdId(cmd, name)
	if param, exists := h.params[id]; exists {
		if v, ok := param.(func() *int); ok {
			return *v()
		}
		return 0
	}
	return 0
}

func (h *Handler) GetBool(cmd *cobra.Command, name string) bool {
	id := h.getCmdId(cmd, name)
	if param, exists := h.params[id]; exists {
		if v, ok := param.(func() *bool); ok {
			return *v()
		}
		return false
	}
	return false
}

func (h *Handler) InitViper() {
	path, err := os.Getwd()
	if err == nil {
		h.viper.AddConfigPath(path)
	}
	cliDataPath, err := constants.CliDataFilePath()
	if err == nil {
		h.viper.AddConfigPath(cliDataPath)
	}

	h.viper.SetConfigName("zcli.config")
	h.viper.SetEnvPrefix("ZEROPS")
	h.viper.AutomaticEnv()

	if err := h.viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", h.viper.ConfigFileUsed())
	}
}

func toSnakeCase(flagName string) string {
	var result string
	for i, r := range flagName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += "_"
		}
		result += string(r)
	}
	return result
}
