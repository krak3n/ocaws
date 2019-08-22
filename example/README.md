# ocaws Example

This example sets up a full end to end server / client flow using OpenCensus to
trace messages from a HTTP Request, to SNS and finally through to an SQS
consumer.

You can run this locally using localstack or connect it to a real AWS SNS/SQS
subscription.

Configuration is handled using environment variables, a full list can be found
below.

By default Jaeger is used as the OpenCensus exporter and is configured to be
running locally. Either start a Jaeger docker container as described here:
https://www.jaegertracing.io/docs/1.13/getting-started/#all-in-one OR use the
provided `docker-compose` which will also start a Localstack: `make compose-up`.

## Running against Localstack

To run against Localstack you'll need `docker` and `docker-compose`. Install
them using your OS's recommended installation method. Once installed run the
following commands in your terminal from this directory.

```
make compose-up
LOCALSTACK=true go run ./
```

You should see something similar to the following output:

```
2019/08/22 11:58:37 Localstack: Created Topic: arn:aws:sns:eu-west-1:000000000000:ocaws
2019/08/22 11:58:37 Localstack: Created Queue: arn:aws:sqs:eu-west-1:000000000000:ocaws
2019/08/22 11:58:37 Localstack: Created Subscription: arn:aws:sns:eu-west-1:000000000000:ocaws:1a866afc-af00-4726-88fa-2c2e882724f4
2019/08/22 11:58:37 Client: Consuming from Queue: http://localhost:4576/queue/ocaws
2019/08/22 11:58:37 main: Server Started: :8080
```

Now the server is running and the client is listening for messages, now you can
send a `POST` request to the server, the request body will become the message
body sent to the SNS topic.

```
curl http://localhost:8080 \
	--header "Content-Type: application/json" \
	--request POST \
	--data '{"foo":"bar","fizz":"buzz"}'
```

You should then see log output similar to the following:

```
2019/08/22 12:05:29 Server: Published message to SNS: arn:aws:sns:eu-west-1:000000000000:ocaws
2019/08/22 12:05:29 Handler: {
  Body: "{\"MessageId\": \"e4d7c618-6cad-4d05-ba70-8ed328c2175b\", \"Type\": \"Notification\", \"Message\": \"{\\\"foo\\\":\\\"bar\\\",\\\"fizz\\\":\\\"buzz\\\"}\", \"TopicArn\": \"arn:aws:sns:eu-west-1:000000000000:ocaws\", \"MessageAttributes\": {\"B3-Span-ID\": {\"Type\": \"String\", \"Value\": \"b5f81a191a4b2885\"}, \"B3-Span-Sampled\": {\"Type\": \"String\", \"Value\": \"1\"}, \"B3-Trace-ID\": {\"Type\": \"String\", \"Value\": \"b130ca24bdf8606296572b5340216c7c\"}, \"Trace-Topic-Name\": {\"Type\": \"String\", \"Value\": \"ocaws\"}}}",
  MD5OfBody: "07325887f6dc9b783c2e03c8ddd909ec",
  MessageId: "bfef8784-d0cd-4db6-a63f-8de7021531ba",
  ReceiptHandle: "bfef8784-d0cd-4db6-a63f-8de7021531ba#b945d00f-6833-434f-ac3a-cc0ec9843a33"
}
```

Once you are done you can shut Localstack down:

```
make compose-down
```

## Configuration

Here are all the environment variables you can set to configure the example:

| Name | Description |
|--------------|---------------------------------------------------------------------------------|
| `LOCALSTACK` | Set true to run against Localstack |
| `SERVER_PORT` | Port for the HTTP server to listen on, defaults to `:8080` |
| `SQS_ENDPOINT` | Endpoint of the SQS API, this gets set explicitly if running against Localstack |
| `SQS_REGION` | SQS AWS Region, defaults to `eu-west-1` |
| `SQS_QUEUE_NAME` | SQS queue name to create when using Localstack, defaults to `ocaws` |
| `SQS_QUEUE_URL` | SQS queue URL to use, when running against Localstack this will be overridden |
| `SNS_ENDPOINT` | Endpoint of the SNS API, this gets set explicitly if running against Localstack |
| `SNS_REGION` | SNS AWS Region, defaults to `eu-west-1` |
| `SNS_TOPIC_NAME` | SNS topic name to create when using Localstack, defaults to `ocaws` |
| `SNS_TOPIC_ARN` | SNS topic ARN to use, when running against Localstack this will be overridden |
| `SNS_SUBSCRIPTION_RAW_MESSAGE_DELIVERY` | Set true if you want to use Raw Message Delivery on your SQS SNS subscription |
| `TRACE_SERVICE` | Service name to use in traces, defaults to `ocaws` |
| `TRACE_JAEGER_AGENT_ENDPOINT` | Jaeger agent endpoint, defaults to `localhost:6831` |
| `TRACE_JAEGER_COLLECTOR_ENDPOINT` | Jaeger collector endpoint, defaults to `http://localhost:14268/api/traces` |
