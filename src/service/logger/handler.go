package logger

import "fmt"

type Config struct {
}

type Handler struct {
	config Config
}

func New(config Config) *Handler {
	return &Handler{
		config: config,
	}
}

func (h *Handler) Info(a ...interface{}) {
	fmt.Println(a...)
}

func (h *Handler) Warning(a ...interface{}) {
	fmt.Println(a...)
}

func (h *Handler) Error(a ...interface{}) {
	fmt.Println(a...)
}

func (h *Handler) Debug(a ...interface{}) {
	fmt.Println(a...)
}
