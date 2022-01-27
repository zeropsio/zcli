package wgquick

import (
	_ "embed"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

//go:embed templates/wg.conf.tmpl
var configTemplate string

func Write(path string, config Config) error {
	err := os.MkdirAll(filepath.Dir(path), 0664)
	if err != nil {
		return err
	}

	tmpl := template.Must(template.New("").Parse(configTemplate))

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tmpl.Execute(f, struct {
		ClientPrivateKey string
		ClientAddress    string
		DnsServers       string
		MTU              string

		ServerPublicKey string
		AllowedIPs      string
		ServerAddress   string
	}{
		ClientPrivateKey: config.ClientPrivateKey,
		AllowedIPs:       config.AllowedIPs.String(),
		DnsServers:       strings.Join(config.DnsServers, ", "),
		ServerAddress:    config.ServerAddress,
		ServerPublicKey:  config.ServerPublicKey,
		ClientAddress:    config.ClientAddress.String(),
		MTU:              strconv.Itoa(config.MTU),
	})

	return err
}
