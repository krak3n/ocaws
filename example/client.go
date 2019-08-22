package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/spf13/viper"
	"go.krak3n.codes/ocaws/ocsqs"
	"go.opencensus.io/trace"
)

// Handler handles a message
type Handler func(context.Context, *sqs.Message) error

// Client consumes messages from SQS
type Client struct {
	sqs    ocsqs.SQSAPI
	semC   chan struct{}
	closeC chan struct{}
	wg     sync.WaitGroup
}

// NewClient constructs a new client to process messages
func NewClient(sqs ocsqs.SQSAPI) *Client {
	return &Client{
		sqs:    sqs,
		semC:   make(chan struct{}, 10),
		closeC: make(chan struct{}),
	}
}

// Consume consumes messages using a semaphore passing the received messages
// into the given handler
func (c *Client) Consume(ctx context.Context, handler Handler) error {
	viper.SetDefault("sqs.queue.max_number_of_messages", 10)

	c.wg.Add(1)
	defer c.wg.Done()

	log.Println("Client: Consuming from Queue:", viper.GetString("sqs.queue.url"))

	for {
		select {
		case <-c.closeC:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			msgs, err := c.sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
				QueueUrl:              aws.String(viper.GetString("sqs.queue.url")),
				MaxNumberOfMessages:   aws.Int64(viper.GetInt64("sqs.queue.max_number_of_messages")),
				MessageAttributeNames: []*string{aws.String("All")},
			})
			if err != nil {
				return err
			}

			if len(msgs.Messages) == 0 {
				continue
			}

			c.wg.Add(len(msgs.Messages))

			for _, msg := range msgs.Messages {
				c.semC <- struct{}{}

				go func(ctx context.Context, msg *sqs.Message) {
					ctx, span := c.sqs.StartSpanFromMessage(ctx, msg)
					defer span.End()

					defer func() {
						<-c.semC
						c.wg.Done()
					}()

					if err := handler(ctx, msg); err != nil {
						log.Println("Client:", err)
					}

					_, err := c.sqs.DeleteMessage(&sqs.DeleteMessageInput{
						QueueUrl:      aws.String(viper.GetString("sqs.queue.url")),
						ReceiptHandle: msg.ReceiptHandle,
					})
					if err != nil {
						log.Println("Client:", err)
					}
				}(ctx, msg)
			}
		}
	}
}

// Stop stops the consumer receiving messages and waits for all messages
// inflight to be processed before exiting
func (c *Client) Stop() {
	close(c.closeC)
	c.wg.Wait()
}

// DefaultHandler is the default handler that conumes a message and simulates
// work by sleeping for a random time
func DefaultHandler(ctx context.Context, msg *sqs.Message) error {
	ctx, span := trace.StartSpan(ctx, "consumer.DefaultHandler")
	defer span.End()

	log.Println("Handler:", msg)
	time.Sleep(time.Second)

	return nil
}
