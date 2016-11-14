package types

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-utils/errors"
)

type x509Loader struct{
	certFilePath, keyFilePath string
}

func NewX509Loader(certFilePath, keyFilePath string) CertsLoader {
	return x509Loader{certFilePath, keyFilePath}
}

func (l x509Loader) LoadCerts(_ string) (*x509.Certificate, *rsa.PrivateKey, error) {
	crt, err := l.parseCertificate(l.certFilePath)
	if err != nil {
		return nil, nil, err
	}

	key, err := l.parsePrivateKey(l.keyFilePath)
	if err != nil {
		return nil, nil, err
	}

	return crt, key, nil
}

func (l x509Loader) parseCertificate(certFilePath string) (*x509.Certificate, error) {
	cf, e := ioutil.ReadFile(l.certFilePath)
	if e != nil {
		return nil, errors.Error("Failed to load certificate file")
	}

	cpb, _ := pem.Decode(cf)
	crt, e := x509.ParseCertificate(cpb.Bytes)

	if e != nil {
		return nil, errors.WrapError(e, "Failed to parse certificate")
	}

	return crt, nil
}

func (l x509Loader) parsePrivateKey(keyFilePath string) (*rsa.PrivateKey, error) {
	kf, e := ioutil.ReadFile(l.keyFilePath)
	if e != nil {
		return nil, errors.Error("Failed to load private key file")
	}

	kpb, _ := pem.Decode(kf)

	key, e := x509.ParsePKCS1PrivateKey(kpb.Bytes)
	if e != nil {
		return nil, errors.WrapError(e, "Failed to parse private key")
	}
	return key, nil
}
