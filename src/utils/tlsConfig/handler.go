package tlsConfig

import (
	"crypto/tls"
	"crypto/x509"
	"errors"

	"github.com/zerops-io/zcli/src/utils/certReader"
)

func CreateTlsConfig(certReader *certReader.Handler) (tlsConfig *tls.Config, err error) {
	tlsConfig = &tls.Config{}
	tlsConfig.InsecureSkipVerify = true
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert

	var crt tls.Certificate
	crt, err = tls.X509KeyPair(certReader.CertData, certReader.KeyData)
	if err != nil {
		return nil, err
	}

	tlsConfig.Certificates = []tls.Certificate{crt}

	var caCertPool *x509.CertPool
	caCertPool, err = x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	if ok := caCertPool.AppendCertsFromPEM(certReader.CaData); !ok {
		return nil, errors.New("error append cert from PEM")
	}
	tlsConfig.RootCAs = caCertPool
	tlsConfig.ClientCAs = caCertPool

	return
}
