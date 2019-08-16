/*Package ocaws provides OpenCensus tracing support for distributed systems
using AWS services such as SQS and SNS. You should have some basic familiarity
with OpenCensus concepts, please see http://opencensus.io for more information.

    go get go.krak3n.codes/ocaws

The clients provided here are designed to be drop in replacements for AWS clients
with little configuration required, you can of course tailor each client to your
tracing needs.

The documentation here assumes you have already setup your exporters for tracing
as described in the OpenCensus documentation, for example:

    trace.RegisterExporter(exporter)


SNS

To publish messages using SNS with span context create a ocsns client and publish
a message as you normally would using PublishWithContext.

    ctx, span := trace.StartSpan(contet.Background(), "my.span/Name")
    defer span.End()

    client := ocsns.New(session)
    client.PublishWithContext(ctx, &sns.PublishInput{...})

See more complete examples in the ocsns documentation: https://godoc.org/go.krak3n.codes/ocaws/ocsns#pkg-examples


SQS

To publish a message directly to an SQS queue create an ocsqs client and use the
new SendMessageContext method to send a message to an SQS with span context.

    ctx, span := trace.StartSpan(contet.Background(), "my.span/Name")
    defer span.End()

    client := ocsqs.New(session)
    client.SendMessageContext(ctx, &sqs.SendMessageInput{...})

To receive messages from SQS and start spans use the StartSpanFromMessage method
which will return you a span based on the messages span context message attributes.

    client := ocsqs.New(session)

    rsp, _ := client.ReceiveMessage(&sqs.ReceiveMessageInput{...})
    for _, msg := range rsp.rsp.Messages {
        ctx, span := sqsClient.StartSpanFromMessage(ctx, msg)
        defer span.End()
        // Do work
    }

See more complete examples in the ocsns documentation: https://godoc.org/go.krak3n.codes/ocaws/ocsqs#pkg-examples
*/
package ocaws // import "go.krak3n.codes/ocaws"
