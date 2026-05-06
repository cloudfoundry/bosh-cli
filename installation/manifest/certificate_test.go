package manifest_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
)

var _ = Describe("Certificate", func() {
	Describe("CACertPool", func() {
		It("returns nil when CA is empty", func() {
			certPool, err := Certificate{CA: ""}.CACertPool()
			Expect(err).ToNot(HaveOccurred())
			Expect(certPool).To(BeNil())
		})

		It("returns a cert pool when a valid PEM CA cert is provided", func() {
			caCert := `-----BEGIN CERTIFICATE-----
MIIDXzCCAkegAwIBAgIJAPerMgLAne5vMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwIBcNMTYwMTE2MDY0NTA0WhgPMjI4OTEwMzAwNjQ1MDRa
MEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJ
bnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQCtSo3KPjnVPzodb6+mNwbCdcpzVop8OmfwJ3ynQtyBEzGaKsAn4tlz
/wfQQrKFHgxqVpqcoxAlWPNMs5+iO2Jst3Gz2+oLcaDyz/EWorw0iF5q1F6+WYHp
EijY20MzaWYMyu4UhhlbJCkSGZSjujh5SFOAXQwWYJXsqjyxA9KaTD6OdH5Kpger
B9D4zogX0We00eouyvvz/sAeDbTshk9sJRGWHNFJr+TjVx2D01alU49liAL94yF6
1eEOEbE50OAhv9RNsRh6O58idaHg30bbMf1yAzcgBvh8CzIHH0BPofoF2pRfztoY
uudZ0ftJjTz4fA2h/7GOVzxemrTjx88vAgMBAAGjUDBOMB0GA1UdDgQWBBQjz5Q2
YW2kBTb4XLqKFZMSBLpi6zAfBgNVHSMEGDAWgBQjz5Q2YW2kBTb4XLqKFZMSBLpi
6zAMBgNVHRMEBTADAQH/MA0GCSqGSIb3DQEBBQUAA4IBAQA/s94M/mSGELHJWIb1
oE0IKHWajBd3Pc8+O1TZRE+ke3q+rZRfcxd2dAjq6zQHJUs2+fs0B3DyT9Wtyyoq
UrRdsgprOdf2Cuw8bMIsCQOvqWKhhdlLTnCi2xaGJawGsIkheuD1n+Il9gRQ2WGy
lACxVngPwjNYxjOE+CUnSZCuAmAfQYzqto3bNPqkgEwb7ueODeOiyhR8SKsH7ySW
QAOSxgrLBblGLWcDF9fjMeYaUnI34pHviCKeVxfgsxDR+Jg11F78sPdYLOF6ipBe
/5qTYucsY20B2EKtlscD0mSYBRwbVrSQt2RYbTCwaibxWUC13VV+YEk0NAv9Mm04
6sKO
-----END CERTIFICATE-----`

			certPool, err := Certificate{CA: caCert}.CACertPool()
			Expect(err).ToNot(HaveOccurred())
			Expect(certPool).ToNot(BeNil())
		})

		It("returns an error when the CA cert is not valid PEM", func() {
			_, err := Certificate{CA: "not-a-valid-pem"}.CACertPool()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Parsing certificate 1: Missing PEM block"))
		})
	})
})
