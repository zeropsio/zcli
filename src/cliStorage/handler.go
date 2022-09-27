package cliStorage

import "github.com/zeropsio/zcli/src/utils/storage"

type Handler struct {
	*storage.Handler[Data]
}

type Data struct {
	ProjectId string
	ServerIp  string
	Token     string
}
