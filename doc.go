/*Package ocaws provides OpenCensus tracing support for distributed systems
using AWS services such as SQS and SNS. You should have some basic familiarity
with OpenCensus concepts, please see http://opencensus.io for more information.

The clients provided here are designed to be drop in replacements for AWS clients
with little configuration required, you can of course tailor each client to your
tracing needs.

The documentation here assumes you have already setup your exporters for tracing
as described in the OpenCensus documentation, for example:

    trace.RegisterExporter(exporter)

*/
package ocaws // import "go.krak3n.codes/ocaws"
