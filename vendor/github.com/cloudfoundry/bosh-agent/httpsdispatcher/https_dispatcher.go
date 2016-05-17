package httpsdispatcher

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type HTTPSDispatcher struct {
	httpServer *http.Server
	mux        *http.ServeMux
	host       string
	listener   net.Listener
	logger     boshlog.Logger
}

type HTTPHandlerFunc func(writer http.ResponseWriter, request *http.Request)

func NewHTTPSDispatcher(baseURL *url.URL, logger boshlog.Logger) *HTTPSDispatcher {
	tlsConfig := &tls.Config{
		// SSLv3 is insecure due to BEAST and POODLE attacks
		MinVersion: tls.VersionTLS10,
		// Both 3DES & RC4 ciphers can be exploited
		// Using Mozilla's "Modern" recommended settings (where they overlap with golang support)
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
		PreferServerCipherSuites: true,
	}
	return NewHTTPSDispatcherWithConfig(tlsConfig, baseURL, logger)
}

func NewHTTPSDispatcherWithConfig(tlsConfig *tls.Config, baseURL *url.URL, logger boshlog.Logger) *HTTPSDispatcher {
	httpServer := &http.Server{
		TLSConfig: tlsConfig,
	}
	mux := http.NewServeMux()
	httpServer.Handler = mux

	return &HTTPSDispatcher{
		httpServer: httpServer,
		mux:        mux,
		host:       baseURL.Host,
		logger:     logger,
	}
}

func (h *HTTPSDispatcher) Start() error {
	tcpListener, err := net.Listen("tcp", h.host)
	if err != nil {
		return bosherr.WrapError(err, "Starting HTTP listener")
	}
	h.listener = tcpListener

	cert, err := tls.LoadX509KeyPair("agent.cert", "agent.key")
	if err != nil {
		return bosherr.WrapError(err, "Loading agent SSL cert")
	}

	// update the server config with the cert
	config := h.httpServer.TLSConfig
	config.NextProtos = []string{"http/1.1"}
	config.Certificates = []tls.Certificate{cert}

	tlsListener := tls.NewListener(tcpListener, config)

	return h.httpServer.Serve(tlsListener)
}

func (h *HTTPSDispatcher) Stop() {
	if h.listener != nil {
		_ = h.listener.Close()
		h.listener = nil
	}
}

func (h *HTTPSDispatcher) AddRoute(route string, handler HTTPHandlerFunc) {
	h.mux.HandleFunc(route, handler)
}
