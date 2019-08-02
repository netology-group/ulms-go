package client

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"net"
	"net/http"
	"time"
)

// HTTP used to make HTTP requests
var HTTP = &WithMetrics{&http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 1 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 1 * time.Second,
	},
}}

// WithMetrics provides metrics for Prometheus
type WithMetrics struct {
	*http.Client
}

// Do request
func (c *WithMetrics) Do(r *http.Request, metrics prometheus.ObserverVec) (*http.Response, error) {
	start := time.Now()
	response, err := c.Client.Do(r)
	if err == nil && metrics != nil {
		metrics.
			With(prometheus.Labels{"code": fmt.Sprintf("%v", response.StatusCode)}).
			Observe(time.Since(start).Seconds())
	}
	return response, err
}
