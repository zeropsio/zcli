package wgquick

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const Template = `
[Interface]
PrivateKey = {{.ClientPrivateKey}}
Address = {{.ClientAddress}}
DNS = {{.DnsServers}}

[Peer]
PublicKey = {{.ServerPublicKey}}
AllowedIPs = {{.AllowedIPs}}
Endpoint = {{.ServerAddress}}
`

func Write(path string, config Config) error {
	err := os.MkdirAll(filepath.Dir(path), 0664)
	if err != nil {
		return err
	}

	tmpl := template.Must(template.New("").Parse(Template))

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tmpl.Execute(f, struct {
		ClientPrivateKey string
		ClientAddress    string
		DnsServers       string

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
	})

	return err
}
