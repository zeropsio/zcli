package certReader

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"strconv"
	"strings"
)

var userIdType = asn1.ObjectIdentifier([]int{1, 2, 3, 1})

type Config struct {
	Token string
}

type Handler struct {
	config Config

	CaData, CertData, KeyData []byte

	UserId string
}

func New(config Config) (h *Handler, err error) {

	h = &Handler{
		config: config,
	}

	tokens := strings.Split(config.Token, ";")
	if len(tokens) != 3 {
		return h, errors.New("wrong token")
	}

	if h.CaData, err = readData(tokens[0]); err != nil {
		return
	}
	if h.KeyData, err = readData(tokens[1]); err != nil {
		return
	}
	if h.CertData, err = readData(tokens[2]); err != nil {
		return
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(h.CertData)
	if !ok {
		return nil, errors.New("failed to parse root certificate")
	}

	block, _ := pem.Decode(h.CertData)
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.New("failed to parse certificate: " + err.Error())
	}

	h.UserId, err = func(cert *x509.Certificate) (string, error) {
		for _, name := range cert.Subject.Names {
			switch {
			case name.Type.Equal(userIdType):
				return name.Value.(string), nil
			}
		}
		return "", errors.New("bad certificate, try contact support")
	}(cert)
	if err != nil {
		return
	}

	return
}

func readData(value string) (data []byte, err error) {
	valueB, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	value = string(valueB)

	if strings.Contains(value, `\n`) && !strings.HasPrefix(value, `"`) {
		value = `"` + value + `"`
	}
	if !strings.HasPrefix(value, `"`) {
		return []byte(value), err
	}
	valueCaUnquote, err := strconv.Unquote(value)
	if err != nil {
		return nil, err
	}
	return []byte(valueCaUnquote), nil
}
