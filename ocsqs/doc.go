/*Package ocsqs provides a drop in replacement for your exisitng SQS client
providing methods for persisiting and creating spans from SQS message
attributes.

    client := ocsqs.New(sqs.New(session))


Raw Message Delivery

SNS allows you to create subscriptions with RawMessageDelivery enabled. If you
have connected your SQS queue to an SNS topic with RawMessageDelivery you must
create the ocsqs.SQS client also with RawMessageDelivery enabled as this affects
how span contexts are retrieved from message attributes.

    client := ocsqs.New(sqs.New(session), ocsqs.WithRawMessageDelivery())

Rember to set the MessageAttributeNames field on ReceiveMessageInput to All
to ensure message attributes are added to the message:

    msgs, err := sqs.ReceiveMessage(&sqs.ReceiveMessageInput{
        QueueUrl:              aws.String("your-queue-url"),
        MessageAttributeNames: []*string{aws.String("All")},
    })

*/
package ocsqs // import "go.krak3n.codes/ocaws/ocsqs"
