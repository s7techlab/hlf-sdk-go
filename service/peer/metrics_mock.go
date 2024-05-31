package peer

import "context"

type EmptyMetricsClient struct{}

func NewEmptyMetricsClient() *EmptyMetricsClient {
	return &EmptyMetricsClient{}
}

func (c *EmptyMetricsClient) GetFabricVersion(ctx context.Context) (string, error) {
	return "", nil
}
