package peer

import "context"

type MockMetricsClient struct{}

func NewMockMetricsClient() *MockMetricsClient {
	return &MockMetricsClient{}
}

func (c *MockMetricsClient) GetFabricVersion(ctx context.Context) (string, error) {
	return "", nil
}
