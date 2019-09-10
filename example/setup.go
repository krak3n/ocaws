package main

import (
	"log"

	"contrib.go.opencensus.io/exporter/jaeger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/spf13/viper"
	"go.krak3n.codes/ocaws/ocsns"
	"go.krak3n.codes/ocaws/ocsqs"
	"go.opencensus.io/trace"
)

// RegisterExporter sets up the Jaeger exporter
func RegisterExporter() {
	viper.SetDefault("trace.service", "ocaws")
	viper.SetDefault("trace.jaeger.agent.endpoint", "localhost:6831")
	viper.SetDefault("trace.jaeger.collector.endpoint", "http://localhost:14268/api/traces")

	e, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint:     viper.GetString("trace.jaeger.agent.endpoint"),
		CollectorEndpoint: viper.GetString("trace.jaeger.collector.endpoint"),
		ServiceName:       viper.GetString("trace.service"),
	})
	if err != nil {
		log.Fatal(err)
	}

	trace.RegisterExporter(e)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})
}

// NewSQS constructs a new SQS client
func NewSQS() *ocsqs.SQS {
	viper.SetDefault("sqs.region", "eu-west-1")

	cfg := &aws.Config{
		Region: aws.String(viper.GetString("sqs.region")),
	}

	if v := viper.GetString("sqs.endpoint"); v != "" {
		cfg.Endpoint = aws.String(v)
	}

	session, err := session.NewSession(cfg)
	if err != nil {
		log.Fatal(err)
	}

	opts := []ocsqs.Option{}
	if viper.GetBool("sns.subscription.raw_message_delivery") {
		log.Println("SQS: Enable Raw Message Delivery")
		opts = append(opts, ocsqs.WithRawMessageDelivery())
	}

	return ocsqs.New(sqs.New(session), opts...)
}

// NewSNS constructs a new SNS client
func NewSNS() *ocsns.SNS {
	viper.SetDefault("sns.region", "eu-west-1")

	cfg := &aws.Config{
		Region: aws.String(viper.GetString("sns.region")),
	}

	if v := viper.GetString("sns.endpoint"); v != "" {
		cfg.Endpoint = aws.String(v)
	}

	session, err := session.NewSession(cfg)
	if err != nil {
		log.Fatal(err)
	}

	return ocsns.New(sns.New(session))
}

// Localstack creates a SNS topic, SQS queue and a SNS > SQS subscription in
// Localstack - due to a bug with Localstack SQS messages published via
// Localstack will not be deleted. It will also update default configurtion for
// SNS topic arn and SQS queue url
func Localstack(s *ocsns.SNS, q *ocsqs.SQS) error {
	// Create topic
	viper.SetDefault("sns.topic.name", "ocaws")
	topic, err := s.CreateTopic(&sns.CreateTopicInput{
		Name: aws.String(viper.GetString("sns.topic.name")),
	})
	if err != nil {
		return err
	}

	viper.Set("sns.topic.arn", *topic.TopicArn)

	// Create Queue
	viper.SetDefault("sqs.queue.name", "ocaws")
	queue, err := q.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(viper.GetString("sqs.queue.name")),
	})
	if err != nil {
		return err
	}

	viper.Set("sqs.queue.url", *queue.QueueUrl)

	// Create Subscription
	// Supports raw message delivery if configured

	attr := map[string]*string{}
	if viper.GetBool("sns.subscription.raw_message_delivery") {
		log.Println("Localstack: Create topic with Raw Message Delivery")
		attr["RawMessageDelivery"] = aws.String("true")
	}

	log.Println("Localstack: Created Topic:", *topic.TopicArn)

	a, err := q.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl:       queue.QueueUrl,
		AttributeNames: []*string{aws.String("All")},
	})
	if err != nil {
		return err
	}

	log.Println("Localstack: Created Queue:", *a.Attributes["QueueArn"])

	sub, err := s.Subscribe(&sns.SubscribeInput{
		TopicArn:   topic.TopicArn,
		Endpoint:   a.Attributes["QueueArn"],
		Protocol:   aws.String("sqs"),
		Attributes: attr,
	})
	if err != nil {
		return err
	}

	log.Println("Localstack: Created Subscription:", *sub.SubscriptionArn)

	return err
}
