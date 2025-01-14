package application

import (
	"crypto/tls"
	"net/http"
)

type HttpTransport struct {
	T         http.RoundTripper
	UserAgent string
}

func (m *HttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", m.UserAgent)
	return m.T.RoundTrip(req)
}

func NewHttpTransport(useragent string) *HttpTransport {
	return &HttpTransport{
		T: &http.Transport{
			TLSClientConfig: &tls.Config{},
		},
		UserAgent: useragent,
	}
}
