package service

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

func newHTTPClient(timeout time.Duration, config ...TimeoutConfig) *http.Client {
	dialTimeout := 30 * time.Second
	tlsHandshakeTimeout := 15 * time.Second
	responseHeaderTimeout := 30 * time.Second

	if len(config) > 0 {
		cfg := config[0]
		if cfg.HTTPTimeout > 0 {
			dialTimeout = time.Duration(cfg.HTTPTimeout) * time.Second
		}
		if cfg.AITLSHandshakeTimeout > 0 {
			tlsHandshakeTimeout = time.Duration(cfg.AITLSHandshakeTimeout) * time.Second
		}
		if cfg.AIResponseHeaderTimeout > 0 {
			responseHeaderTimeout = time.Duration(cfg.AIResponseHeaderTimeout) * time.Second
		}
	}

	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   dialTimeout,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   tlsHandshakeTimeout,
			ResponseHeaderTimeout: responseHeaderTimeout,
			IdleConnTimeout:       90 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
		},
	}
}
