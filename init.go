// Package to interact programmatically with the school register ClasseViva by Spaggiari
package cvvapi

import (
	"net/http"
	"time"
)

// Configuration structure accessible through chainable setters
type PkgConfig struct {
	cvvHostname string
	requestMiddlewares []HttpMiddleware
	httpClientProvider func () *http.Client

	sessionCheckTimeout time.Duration
}

var config = PkgConfig{}

func init() {
	config.cvvHostname = "web.spaggiari.eu"
	config.sessionCheckTimeout, _ = time.ParseDuration("20m")
	
	sharedHttpClient := &http.Client{
		Transport: &http.Transport{},
	}

	config.httpClientProvider = func() *http.Client {
		return sharedHttpClient
	}
}

// Expose a handle to the module configuration
func Config() *PkgConfig {
	return &config
}

// Set the configured hostname. Default "web.spaggiari.eu"
func (o *PkgConfig) SetHostname(hostname string) *PkgConfig {
	config.cvvHostname = hostname
	return o
}

// Set the http client provider function. Default shared client
func (o *PkgConfig) SetClientProvider(provider func() *http.Client) *PkgConfig {
	config.httpClientProvider = provider
	return o
}

// Set the timeout of the cvv session. Default 20 minutes
func (o *PkgConfig) SetSessionCheckTimeout(timeout time.Duration) *PkgConfig {
	config.sessionCheckTimeout = timeout
	return o
}

// Get an http client from the provider function
func (o *PkgConfig) GetClient() *http.Client {
	return config.httpClientProvider()
}

// Add a request middleware to the middleware pipeline. 
// Middlewares are executed in the order in which they are added
func (o *PkgConfig) Use(mw HttpMiddleware) *PkgConfig {
	o.requestMiddlewares = append(o.requestMiddlewares, mw)
	httpPipeline = buildHttpPipeline()
	return o
}

func buildHttpPipeline() HttpHandler {
	var pipeline HttpHandler = doRequest
	for i := len(config.requestMiddlewares) - 1; i >= 0; i-- {
		fn := config.requestMiddlewares[i]
		next := pipeline
		pipeline = func(r *http.Request) (*http.Response, error) {
			return fn(r, next)
		}
	}
	return pipeline
}
