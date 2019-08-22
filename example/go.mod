module go.krak3n.codes/ocaws/example

go 1.12

replace go.krak3n.codes/ocaws => ./../

require (
	contrib.go.opencensus.io/exporter/jaeger v0.1.0
	github.com/aws/aws-sdk-go v1.23.5
	github.com/spf13/viper v1.4.0
	go.krak3n.codes/ocaws v0.0.0-00010101000000-000000000000
	go.opencensus.io v0.22.0
)
