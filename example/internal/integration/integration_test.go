package integration

import (
	"context"
	"testing"

	testmanager "github.com/nais/tester/lua"
)

func TestIntegration(t *testing.T) {
	mgr, err := TestRunner(false)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := mgr.Run(ctx, "./testdata", testmanager.NewTestReporter(t)); err != nil {
		t.Fatal(err)
	}
}
