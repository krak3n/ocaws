package awsoc

import (
	"os"
	"testing"

	"go.krak3n.codes/awsoc/awsoctest"
	"go.opencensus.io/trace"
)

func TestMain(m *testing.M) {
	trace.ApplyConfig(trace.Config{
		IDGenerator: awsoctest.NewTestIDGenerator(),
	})

	os.Exit(m.Run())
}
