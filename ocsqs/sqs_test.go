package ocsqs

import (
	"os"
	"testing"

	"go.krak3n.codes/ocaws/ocawstest"
	"go.opencensus.io/trace"
)

func TestMain(m *testing.M) {
	trace.ApplyConfig(trace.Config{
		IDGenerator: ocawstest.NewTestIDGenerator(),
	})

	os.Exit(m.Run())
}
