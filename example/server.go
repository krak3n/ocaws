package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/spf13/viper"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

// A Publisher publishes messages to an SNS topic
type Publisher interface {
	PublishWithContext(aws.Context, *sns.PublishInput, ...request.Option) (*sns.PublishOutput, error)
}

// NewServer constructs a new http.Server
// This will publish a message to a SNS topic on http requests, the request body
// will be the message body, only POST requests are supported
func NewServer(publisher Publisher) *http.Server {
	viper.SetDefault("server.port", ":8080")

	srv := &http.Server{
		Addr: viper.GetString("server.port"),
		Handler: &ochttp.Handler{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					log.Println("Server: invalid method:", r.Method)
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}

				b, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Println("Server:", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				defer r.Body.Close()

				_, err = publisher.PublishWithContext(r.Context(), &sns.PublishInput{
					TopicArn: aws.String(viper.GetString("sns.topic.arn")),
					Message:  aws.String(string(b)),
				})
				if err != nil {
					log.Println("Server:", err)
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					log.Println("Server: Published message to SNS:", viper.GetString("sns.topic.arn"))
					w.WriteHeader(http.StatusOK)
				}
			}),
			IsPublicEndpoint: true,
			StartOptions: trace.StartOptions{
				Sampler: trace.AlwaysSample(),
			},
		},
	}

	return srv
}
