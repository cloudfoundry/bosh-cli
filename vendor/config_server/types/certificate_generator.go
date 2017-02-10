package types

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/cloudfoundry/bosh-utils/errors"
)

type CertificateGenerator struct {
	loader CertsLoader
}

type CertResponse struct {
	Certificate string `json:"certificate" yaml:"certificate"`
	PrivateKey  string `json:"private_key" yaml:"private_key"`
	CA          string `json:"ca"          yaml:"ca"`
}

type CertParams struct {
	CommonName      string
	AlternativeName []string
	CA              string // todo
	ExtKeyUsage     []x509.ExtKeyUsage
}

func NewCertificateGenerator(loader CertsLoader) CertificateGenerator {
	return CertificateGenerator{loader: loader}
}

func (cfg CertificateGenerator) Generate(parameters interface{}) (interface{}, error) {
	params := parameters.(map[interface{}]interface{}) // todo map style
	commonName := params["common_name"].(string)
	alternativeNames := []string{}
	ca := ""

	if params["alternative_names"] != nil {
		for _, altName := range params["alternative_names"].([]interface{}) {
			alternativeNames = append(alternativeNames, altName.(string))
		}
	}

	if _, ok := params["ca"]; ok {
		ca = params["ca"].(string)
	}

	extKeyUsages := []x509.ExtKeyUsage{}

	if _, ok := params["ext_key_usage"]; ok {
		for _, extKeyUsage := range params["ext_key_usage"].([]interface{}) {
			extKeyUsageString := extKeyUsage.(string)

			switch extKeyUsageString {
			case "client_auth":
				extKeyUsages = append(extKeyUsages, x509.ExtKeyUsageClientAuth)
			case "server_auth":
				extKeyUsages = append(extKeyUsages, x509.ExtKeyUsageServerAuth)
			default:
				return nil, fmt.Errorf("Unsupported extended key usage value: %s", extKeyUsageString)
			}
		}
	} else {
		extKeyUsages = append(extKeyUsages, x509.ExtKeyUsageServerAuth)
	}

	cParams := CertParams{
		CommonName:      commonName,
		AlternativeName: alternativeNames,
		CA:              ca,
		ExtKeyUsage:     extKeyUsages,
	}

	if len(cParams.CA) > 0 {
		return cfg.generateCert(cParams)
	}

	return cfg.generateCACert(cParams)
}

func (cfg CertificateGenerator) generateCert(cParams CertParams) (CertResponse, error) {
	var certResponse CertResponse

	if cfg.loader == nil {
		panic("Expected CertificateGenerator to have Loader set")
	}

	rootCA, rootCAKey, err := cfg.loader.LoadCerts(cParams.CA)
	if err != nil {
		return certResponse, errors.WrapError(err, "Loading certificates")
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return certResponse, errors.WrapError(err, "Generating Serial Number")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return certResponse, errors.WrapError(err, "Generating Key")
	}

	now := time.Now()
	notAfter := now.Add(365 * 24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:      []string{"USA"},
			Organization: []string{"Cloud Foundry"},
			CommonName:   cParams.CommonName,
		},
		NotBefore:             now,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment | x509.KeyUsageKeyAgreement,
		ExtKeyUsage:           cParams.ExtKeyUsage,
		BasicConstraintsValid: true,
		MaxPathLen:            1,
		IsCA:                  false,
	}

	for _, altName := range cParams.AlternativeName {
		possibleIP := net.ParseIP(altName)
		if possibleIP == nil {
			template.DNSNames = append(template.DNSNames, altName)
		} else {
			template.IPAddresses = append(template.IPAddresses, possibleIP)
		}
	}

	template.SubjectKeyId, err = generateSubjectKeyID(&privateKey.PublicKey)
	if err != nil {
		return certResponse, errors.WrapError(err, "Generating Subject Key ID")
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, rootCA, &privateKey.PublicKey, rootCAKey)
	if err != nil {
		return certResponse, errors.WrapError(err, "Generating Certificate")
	}

	encodedCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	encodedPrivatekey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	encodedRootCACert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootCA.Raw})

	certResponse = CertResponse{
		Certificate: string(encodedCert),
		PrivateKey:  string(encodedPrivatekey),
		CA:          string(encodedRootCACert),
	}

	return certResponse, nil
}

func (cfg CertificateGenerator) generateCACert(cParams CertParams) (CertResponse, error) {
	var certResponse CertResponse

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return certResponse, errors.WrapError(err, "Generating Serial Number")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return certResponse, errors.WrapError(err, "Generating Key")
	}

	now := time.Now()
	notAfter := now.Add(365 * 24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:      []string{"USA"},
			Organization: []string{"Cloud Foundry"},
			CommonName:   cParams.CommonName,
		},
		NotBefore:             now,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		BasicConstraintsValid: true,
		MaxPathLenZero:        true,
		IsCA:                  true,
	}

	template.SubjectKeyId, err = generateSubjectKeyID(&privateKey.PublicKey)
	if err != nil {
		return certResponse, errors.WrapError(err, "Generating Subject Key ID")
	}
	template.AuthorityKeyId = template.SubjectKeyId

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return certResponse, errors.WrapError(err, "Generating CA certificate")
	}

	encodedCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	encodedPrivatekey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	encodedRootCACert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	certResponse = CertResponse{
		Certificate: string(encodedCert),
		PrivateKey:  string(encodedPrivatekey),
		CA:          string(encodedRootCACert),
	}

	return certResponse, nil
}

// rsaPublicKey reflects the ASN.1 structure of a PKCS#1 public key.
type rsaPublicKey struct {
	N *big.Int
	E int
}

// GenerateSubjectKeyID generates SubjectKeyId used in Certificate
// Id is 160-bit SHA-1 hash of the value of the BIT STRING subjectPublicKey
func generateSubjectKeyID(pub crypto.PublicKey) ([]byte, error) {
	var pubBytes []byte
	var err error
	switch pub := pub.(type) {
	case *rsa.PublicKey:
		pubBytes, err = asn1.Marshal(rsaPublicKey{
			N: pub.N,
			E: pub.E,
		})
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.Error("only RSA public key is supported")
	}

	hash := sha1.Sum(pubBytes)

	return hash[:], nil
}
