/*Package ocsqs provides a drop in replacement for your exisitng SQS client
* providing methods for persisiting and creating spans from SQS message
* attributes.

    client := ocsqs.New(sqs.New(session))


Raw Message Delivery

SNS allows you to create subscriptions with RawMessageDelivery enabled. If you
have connected your SQS queue to an SNS topic with RawMessageDelivery you must
create the ocsqs.SQS client also with RawMessageDelivery enabled as this affects
how span contexts are retrieved from message attributes.

    client := ocsqs.New(sqs.New(session), ocsqs.WithRawMessageDelivery())

*/
package ocsqs // import "go.krak3n.codes/ocaws/ocsqs"
