package peer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/common/expfmt"
)

const (
	metricsPath = "/metrics"
	reqTimeout  = time.Second * time.Duration(2)
)

var ErrOSNConnection = errors.New("peer operations address is unreachable")

type (
	Metrics interface {
		GetFabricVersion(ctx context.Context) (string, error)
	}

	MetricsClient struct {
		coreOperationsListenAddr string
	}
)

func NewMetricsClient(coreOperationsListenAddr string) (*MetricsClient, error) {
	// ping request
	ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
	defer cancel()

	metricsClient := &MetricsClient{
		coreOperationsListenAddr: coreOperationsListenAddr,
	}

	reqBody, err := metricsClient.httpMetricsReq(ctx, metricsClient.Url())
	if err != nil {
		return nil, fmt.Errorf("peer operations address ping: %w", err)
	}

	defer func() { _ = reqBody.Close() }()

	return metricsClient, nil
}

func (c *MetricsClient) Url() string {
	url := fmt.Sprintf("%s%s", c.coreOperationsListenAddr, metricsPath)
	if !strings.HasPrefix(url, `http://`) && !strings.HasPrefix(url, `https://`) {
		url = `http://` + url
	}

	return url
}

func (c *MetricsClient) GetFabricVersion(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, reqTimeout)
	defer cancel()

	reqBody, err := c.httpMetricsReq(ctx, c.Url())
	if err != nil {
		return "", fmt.Errorf("request building: %w", err)
	}

	defer func() { _ = reqBody.Close() }()

	return c.fabricVersionFromMetrics(reqBody)
}

func (c *MetricsClient) httpMetricsReq(ctx context.Context, addr string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("request building: %w", err)
	}
	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w. reason: %v", ErrOSNConnection, err.Error())
	}

	return res.Body, nil
}

func (c *MetricsClient) fabricVersionFromMetrics(in io.ReadCloser) (string, error) {
	tp := expfmt.TextParser{}

	mf, err := tp.TextToMetricFamilies(in)
	if err != nil {
		return "", fmt.Errorf("parsing metrics: %w", err)
	}

	fv, ok := mf["fabric_version"]
	if !ok {
		return "", fmt.Errorf("no 'fabric_version' key in response: %w", err)
	}

	if len(fv.Metric) == 0 {
		return "", fmt.Errorf("no 'fabric_version' metrics in response: %w", err)
	}
	if len(fv.Metric[0].Label) == 0 {
		return "", fmt.Errorf("no 'fabric_version' metrics label in response: %w", err)
	}

	return *fv.Metric[0].Label[0].Value, nil
}
