package params

import (
	"fmt"
	"os"

	"github.com/zeropsio/zcli/src/constants"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func (h *Handler) RegisterString(cmd *cobra.Command, name, defaultValue, description string) {
	var paramValue string

	cmd.Flags().StringVar(&paramValue, name, defaultValue, description)

	h.params[h.getCmdId(cmd, name)] = func() *string {
		if cmd.Flags().Lookup(name).Changed {
			return &paramValue
		}
		if h.viper.GetString(name) != "" {
			v := h.viper.GetString(name)
			return &v
		}
		return &paramValue
	}
}

func (h *Handler) RegisterBool(cmd *cobra.Command, name string, defaultValue bool, description string) {
	var paramValue bool

	cmd.Flags().BoolVar(&paramValue, name, defaultValue, description)

	h.params[h.getCmdId(cmd, name)] = func() *bool {
		if cmd.Flags().Lookup(name).Changed {
			return &paramValue
		}
		if h.viper.GetBool(name) != false {
			v := h.viper.GetBool(name)
			return &v
		}
		return &paramValue
	}
}

func (h *Handler) RegisterInt(cmd *cobra.Command, name string, defaultValue int, description string) {
	var paramValue int

	cmd.Flags().IntVar(&paramValue, name, defaultValue, description)

	h.params[h.getCmdId(cmd, name)] = func() *int {
		if cmd.Flags().Lookup(name).Changed {
			return &paramValue
		}
		if h.viper.GetInt(name) != 0 {
			v := h.viper.GetInt(name)
			return &v
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

func (h *Handler) InitViper() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	cliDataPath, err := constants.CliDataFilePath()
	if err != nil {
		return err
	}
	h.viper.AddConfigPath(path)
	h.viper.AddConfigPath(cliDataPath)
	h.viper.SetConfigName("zcli.config")
	h.viper.SetEnvPrefix("ZEROPS")
	h.viper.AutomaticEnv()

	if err := h.viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", h.viper.ConfigFileUsed())
	}

	return nil
}
