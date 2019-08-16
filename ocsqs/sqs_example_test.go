package ocsqs_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.krak3n.codes/ocaws"
	"go.krak3n.codes/ocaws/ocawstest"
	"go.krak3n.codes/ocaws/ocsqs"
	"go.krak3n.codes/ocaws/propagation/b3"
	"go.opencensus.io/trace"
)

func ExampleSQS_SendMessageContext() {
	cfg := &aws.Config{
		Region: aws.String("eu-west-1"),
	}

	if v := os.Getenv("AWS_DEFAULT_REGION"); v != "" {
		cfg.Region = aws.String(v)
	}

	if v := os.Getenv("AWS_SQS_ENDPOINT"); v != "" {
		cfg.Endpoint = aws.String(v)
	}

	session, err := session.NewSession(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx, span := trace.StartSpan(context.Background(), "sqs/ExampleSQS_SendMessageContext")
	defer span.End()

	// Create SNS Client
	c := ocsqs.New(sqs.New(session))

	// Create Topic
	q, err := c.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String("foo"),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Publish message with span context message attributes
	in := &sqs.SendMessageInput{
		QueueUrl:    q.QueueUrl,
		MessageBody: aws.String(`{"foo":"bar"}`),
	}

	if _, err := c.SendMessageContext(ctx, in); err != nil {
		log.Fatal(err)
	}

	fmt.Println("TraceID:", *in.MessageAttributes[b3.TraceIDKey].StringValue)
	fmt.Println("SpanID:", *in.MessageAttributes[b3.SpanIDKey].StringValue)
	fmt.Println("Span Sampled:", *in.MessageAttributes[b3.SpanSampledKey].StringValue)
	fmt.Println("Trace Queue URL:", *in.MessageAttributes[ocaws.TraceQueueURL].StringValue)

	// Output:
	// TraceID: 616263646566676869676b6c6d6e6f71
	// SpanID: 6162636465666768
	// Span Sampled: 0
	// Trace Queue URL: http://localhost:4576/queue/foo
}

func ExampleSQS_StartSpanFromMessage() {
	// Create a message with trace attributes, publish a message via SNS or SQS
	msg := &sqs.Message{
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			b3.TraceIDKey: {
				DataType:    aws.String("String"),
				StringValue: aws.String(ocawstest.DefaultTraceID.String()),
			},
			b3.SpanIDKey: {
				DataType:    aws.String("String"),
				StringValue: aws.String(ocawstest.DefaultSpanID.String()),
			},
			b3.SpanSampledKey: {
				DataType:    aws.String("String"),
				StringValue: aws.String("0"),
			},
		},
	}

	session, err := session.NewSession(&aws.Config{})
	if err != nil {
		log.Fatal(err)
	}

	c := ocsqs.New(sqs.New(session))

	ctx := context.Background()
	ctx, span := c.StartSpanFromMessage(ctx, msg)
	defer span.End()

	if span != nil {
		sc := span.SpanContext()
		fmt.Println("TraceID:", sc.TraceID.String())
		fmt.Println("SpanID:", sc.SpanID.String())
		fmt.Println("Span Sampled:", sc.IsSampled())
	}

	// Output:
	// TraceID: 616263646566676869676b6c6d6e6f71
	// SpanID: 6162636465666768
	// Span Sampled: false
}
