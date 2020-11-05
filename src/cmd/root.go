package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	paramsPackage "github.com/zerops-io/zcli/src/utils/params"
)

var (
	params *paramsPackage.Handler
)

var BuiltinToken string

func ExecuteCmd() error {
	params = paramsPackage.New()

	rootCmd := &cobra.Command{
		Use: "zcli",
	}

	params.RegisterPersistentString(rootCmd, "restApiAddress", "https://app.zerops.dev", "address of rest api")
	params.RegisterPersistentString(rootCmd, "grpcApiAddress", "app.zerops.dev:20902", "address of grpc api")
	params.RegisterPersistentString(rootCmd, "vpnApiAddress", "vpn.app.zerops.dev", "address of vpn api")
	params.RegisterPersistentString(rootCmd, "caCertificate", defaultZeropsCACertificate, "certificate of Zerops certificate authority used for tls encrypted communication via gRPC")

	rootCmd.AddCommand(deployCmd())
	rootCmd.AddCommand(pushCmd())
	rootCmd.AddCommand(vpnCmd())
	rootCmd.AddCommand(loginCmd())
	rootCmd.AddCommand(logCmd())
	rootCmd.AddCommand(daemonCmd())

	err := params.InitViper()
	if err != nil {
		return err
	}

	err = rootCmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func regSignals(contextCancel func()) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println("\n", "signal:", sig)
		contextCancel()
	}()
}

const defaultZeropsCACertificate = `-----BEGIN CERTIFICATE-----
MIID6jCCAtKgAwIBAgICB+QwDQYJKoZIhvcNAQELBQAwgYkxCzAJBgNVBAYTAkNa
MQ8wDQYDVQQIEwZQcmFndWUxDzANBgNVBAcTBlByYWd1ZTEZMBcGA1UECRMQU29k
b21rb3ZhIDE1NzkvNTEOMAwGA1UEERMFMTAyMDAxGTAXBgNVBAoTEFZTSG9zdGlu
ZyBzLnIuby4xEjAQBgNVBAMTCXplcm9wcy5pbzAgFw0xOTA3MjcxMzU4MDVaGA8y
MTIwMDcyNzEzNTgwNVowgYkxCzAJBgNVBAYTAkNaMQ8wDQYDVQQIEwZQcmFndWUx
DzANBgNVBAcTBlByYWd1ZTEZMBcGA1UECRMQU29kb21rb3ZhIDE1NzkvNTEOMAwG
A1UEERMFMTAyMDAxGTAXBgNVBAoTEFZTSG9zdGluZyBzLnIuby4xEjAQBgNVBAMT
CXplcm9wcy5pbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMQOFw5F
epu+90na1n6K/G5JSBgBtrBF/V9BvTs7+C716vquw/EDTAhhboL8EcvbQaGUuTvL
Co38GD8zv9A2k8M3l8iWQmEsHhvZv+21Cn/P5mwPxHnJm6STrkc0s3fQY/6q7zFU
ELF/4ICmjv9IVW4t13h4aD5J24va3tC7rmZchgrR2+wsrYq7DuUpt0sGk56Oh5Wp
FR34RvDjIdWc77wHFVfSRmGOPxLV5ZbfvQixN2ZNWQjFz0FQjgNrPK1fFBRMBf0d
QzeYz8ku0315mcKwWc/gSSAQBzs6bIv8z0nJeTgGKTJVDfPtb0kzGISC9usOvIuh
cfDrrBwPRqOu1l8CAwEAAaNYMFYwDgYDVR0PAQH/BAQDAgKEMB0GA1UdJQQWMBQG
CCsGAQUFBwMCBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MBQGA1UdEQQNMAuC
CXplcm9wcy5pbzANBgkqhkiG9w0BAQsFAAOCAQEAeEbQlIRd1KDZJxP++rU3J28t
+pLn4TLwR6fUBaSI4plrZkO/5wibhfKYuHs2ikUtZ4PRGbDjx6qnTcRhBASDTCOA
avnsgsrxlE2YDiSyHy7gPzHv/LYMIzM4dXIHlfMk2E73bHAnG5BZVgnMrCxJ2nIl
z04AOUJN5usLANI2l+ExSsg+mall6kggTi84ImnsVCevb2pz4bQ3mDHLf86Emu3o
pvSWmWdIc3A71HtYSZjxbobiCsriWQmwXPcUPHPveuzULraoAcQSefAZ9/dxLNFZ
CjdL1udHsEL8XqJgvzBkrmRWHDRBWabdsc6bCyxcqzFGmZ4LUvSxLT6Smgj7xA==
-----END CERTIFICATE-----
`
