package creds

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
)

// Config holds the location info for certificates and host name
type Config struct {
	CA   string
	Cert string
	Key  string
	Host string
}

// HasCerts returns true when CA, Cert, Key are not empty
func (c *Config) HasCerts() bool {
	return c.CA != "" && c.Cert != "" && c.Key != ""
}

// HasHost returns true if Host is not empty
func (c *Config) HasHost() bool {
	return c.Host != ""
}

// GetServerCredentials creates transport credentials for server
func GetServerCredentials(c *Config) (credentials.TransportCredentials, error) {
	return getCredentials(c, false)
}

// GetClientCredentials creates transport credentials for client
func GetClientCredentials(c *Config) (credentials.TransportCredentials, error) {
	return getCredentials(c, true)
}

func getCredentials(c *Config, forClient bool) (credentials.TransportCredentials, error) {
	if c == nil || c.CA == "" || c.Cert == "" || c.Key == "" {
		return nil, errors.New("missing certificates")
	}

	ca, err := ioutil.ReadFile(c.CA)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading ca file: %s", c.CA)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(ca) {
		return nil, errors.New("error adding client CA")
	}

	cerPair, err := tls.LoadX509KeyPair(c.Cert, c.Key)
	if err != nil {
		return nil, err
	}

	var config *tls.Config
	if forClient {
		// client
		config = &tls.Config{
			Certificates: []tls.Certificate{cerPair},
			RootCAs:      certPool,
			ServerName:   c.Host,
		}

	} else {
		// server
		config = &tls.Config{
			Certificates: []tls.Certificate{cerPair},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
		}
	}

	return credentials.NewTLS(config), nil
}
