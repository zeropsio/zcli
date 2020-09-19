package params

import (
	"fmt"
	"os"

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

type Option func(cfg *optionConfig)

func Persistent() Option {
	return func(cfg *optionConfig) {
		cfg.persistent = true
	}
}

func FromTempData(f func() interface{}) Option {
	return func(cfg *optionConfig) {
		cfg.loadFromTempData = f
	}
}

type optionConfig struct {
	persistent       bool
	loadFromTempData func() interface{}
}

func (h *Handler) RegisterString(cmd *cobra.Command, name, defaultValue, description string, options ...Option) {
	var paramValue string

	cfg := &optionConfig{}
	for _, o := range options {
		o(cfg)
	}

	if cfg.persistent {
		cmd.PersistentFlags().StringVar(&paramValue, name, "", description)
		h.viper.BindPFlags(cmd.PersistentFlags())
	} else {
		cmd.Flags().StringVar(&paramValue, name, "", description)
	}

	h.params[name] = func() string {
		if paramValue != "" {
			return paramValue
		}
		if cfg.loadFromTempData != nil {
			value := cfg.loadFromTempData()
			if v, ok := value.(string); ok {
				return v
			}
		}
		if h.viper.GetString(name) != "" {
			return h.viper.GetString(name)
		}

		return defaultValue
	}
}

func (h *Handler) GetString(name string) string {
	if param, exists := h.params[name]; exists {
		if v, ok := param.(func() string); ok {
			return v()
		}
		return ""
	}
	return ""
}

func (h *Handler) InitViper() error {

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	h.viper.AddConfigPath(path)
	h.viper.SetConfigName("zcli.config")
	h.viper.AutomaticEnv()

	if err := h.viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", h.viper.ConfigFileUsed())
	}

	return nil
}
