package wgquick

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

type Configurator struct {
	configPath    string
	upCommand     *exec.Cmd
	downCommand   *exec.Cmd
	additionalDns []string
}

func (c Configurator) Up(ifName string, config Config) error {
	binary := c.upCommand.Path

	_, err := exec.LookPath(binary)
	if err != nil {
		return err
	}

	path := filepath.Join(c.configPath, ifName+".conf")

	config.DnsServers = append(config.DnsServers, c.additionalDns...)

	err = Write(path, config)
	if err != nil {
		return err
	}

	c.upCommand.Args = append(c.upCommand.Args, path)

	output, err := c.upCommand.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
	}
	return err
}

func (c Configurator) Down(ifName string) error {
	binary := c.downCommand.Path

	_, err := exec.LookPath(binary)
	if err != nil {
		return err
	}

	c.downCommand.Args = append(c.downCommand.Args, ifName)

	output, err := c.downCommand.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
	}
	return err
}
