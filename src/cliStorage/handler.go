package cliStorage

import "github.com/zerops-io/zcli/src/utils/storage"

type Handler = storage.Handler[Data]

type Data struct {
	ProjectId string
	ServerIp  string
	Token     string
}
