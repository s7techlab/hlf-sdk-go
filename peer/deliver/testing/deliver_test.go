package testing

import (
	"context"
	"testing"
	"time"
)

func TestNewDeliverClient(t *testing.T) {
	dc, err := NewDeliverClient("../testdata/blocks", false)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = dc.Deliver(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
