package scutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"

	"github.com/zerops-io/zcli/src/scutiljson"
)

const BinaryLocation = "/usr/sbin/scutil"

type Routes struct {
	DestinationAddress net.IP
	SubnetMask         net.IP
}
type ServiceIPv4 struct {
	ARPResolvedHardwareAddress string
	ARPResolvedIPAddress       net.IP
	AdditionalRoutes           []Routes
	Addresses                  []net.IP
	ConfirmedInterfaceName     string
	InterfaceName              string
	NetworkSignature           string
	Router                     net.IP
	SubnetMasks                []net.IP
}

type ServiceDns struct {
	DomainName      string
	ServerAddresses []net.IP
}

type NetworkGlobalIPv4 struct {
	PrimaryInterface string
	PrimaryService   string
	Router           net.IP
}

func UnmarshalKey(key string, output interface{}) error {
	jsonOutput := bytes.NewBuffer(nil)
	if err := GetKey(key, jsonOutput); err != nil {
		return err
	}
	fmt.Println(jsonOutput.String())
	return json.NewDecoder(jsonOutput).Decode(output)
}

type KeyValue struct {
	Key    string
	Value  string
	Delete bool
}

func IPsToArrayValue(in ...net.IP) string {
	r := strings.Builder{}
	r.WriteString("*")
	for _, i := range in {
		r.WriteString(" ")
		r.WriteString(i.String())
	}
	return r.String()
}
func ChangeKey(key string, values ...KeyValue) error {
	return MoveKey(key, key, values...)
}

func KeyExists(key string) bool {
	input := bytes.NewBuffer(nil)
	output := bytes.NewBuffer(nil)
	fmt.Fprintf(input, "get %s\n", key)
	cmd := exec.Command(BinaryLocation)
	cmd.Stdin = input
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		return false
	}
	fmt.Println("Exists: ", key)
	fmt.Println(output.String())
	fmt.Println("===========")
	if strings.TrimSpace(output.String()) == "No such key" {
		return false
	}
	return true

}

func MoveKey(oldKey, key string, values ...KeyValue) error {
	input := bytes.NewBuffer(nil)
	output := bytes.NewBuffer(nil)
	input.WriteString("d.init\n")
	fmt.Fprintf(input, "get %s\n", oldKey)
	for _, value := range values {
		if value.Delete {
			fmt.Fprintf(input, "d.remove %s\n", value.Key)
		} else {
			fmt.Fprintf(input, "d.add %s %s\n", value.Key, value.Value)
		}
	}
	fmt.Fprintf(input, "set %s\n", key)
	cmd := exec.Command(BinaryLocation)
	cmd.Stdin = input
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Println(output.String())
	return nil
}

func RemoveKey(key string) error {
	input := bytes.NewBuffer(nil)
	output := bytes.NewBuffer(nil)
	fmt.Fprintf(input, "remove %s\n", key)
	cmd := exec.Command(BinaryLocation)
	cmd.Stdin = input
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Println(output.String())
	return nil
}

func GetKey(key string, writer io.Writer) error {
	output := bytes.NewBuffer(nil)
	input := bytes.NewBuffer([]byte("show " + key))
	cmd := exec.Command(BinaryLocation)
	cmd.Stdin = input
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		return err
	}
	return scutiljson.JSONEncode(output, writer)
}
