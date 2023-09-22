package client

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

type PrometheusV1APIClient struct {
	client v1.API
}

func NewPrometheusV1APIClient(url string) (*PrometheusV1APIClient, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
		RoundTripper: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
	})

	if err != nil {
		logrus.Errorf("Error creating client: %v\n", err)
		return nil, err
	}

	v1api := v1.NewAPI(client)
	return &PrometheusV1APIClient{client: v1api}, nil
}

// e.g. if you have a metric like
// up{hostname="my_vps_id", region="US", location="LA"} 1
// you can get region value or location value of certain hostname from here
// also you can get all the hostname here
func (c *PrometheusV1APIClient) GetStaticLabelValues(query string, target_label string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, _, err := c.client.LabelValues(ctx, target_label, []string{query}, time.Now().Add(-time.Minute*1), time.Now())
	if err != nil {
		logrus.Errorf("Error querying Prometheus: %v", err)
		return nil, err
	}

	labels := make([]string, 0)
	for _, label := range result {
		labels = append(labels, string(label))
	}

	return labels, nil
}

func (c *PrometheusV1APIClient) Query(query string) (model.Value, v1.Warnings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := c.client.Query(ctx, query, time.Now())
	return result, warnings, err
}
