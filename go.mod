module go.krak3n.codes/ocaws

go 1.12

replace go.krak3n.codes/ocaws => ./

require (
	contrib.go.opencensus.io/exporter/jaeger v0.1.0 // indirect
	github.com/aws/aws-sdk-go v1.22.2
	github.com/spf13/viper v1.4.0 // indirect
	github.com/stretchr/testify v1.2.2
	go.opencensus.io v0.22.0
)
