# ocaws - Go OpenCensus AWS Integration

[![Go 1.12+][goversion-image]][goversion-url]
[![Documentation][godoc-image]][godoc-url]
[![Status][circle-image]][circle-url]
[![Coverage][cover-image]][cover-url]
[![Report][report-image]][report-url]

Provides Opencensus Tracing integrations for AWS SQS/SNS via the AWS Go SDK
allowing you to easily propagate spans across SQS/SNS messages.

![jaeger tracing example][jaeger]

# Installation

```
go get go.krak3n.codes/ocaws
```

# Documentation

Please see the [![Documentation][godoc-image]][godoc-url] for code examples and
documentation.

# Example

Please see the [example](example) directory for a full end to end example code.

# Contributing

Please refer to contribution guidelines for submitting patches and additions. In general, we follow the "fork-and-pull" Git workflow.

 1. **Fork** the repo on GitHub
 2. **Clone** the project to your own machine
 3. **Commit** changes to your own branch
 4. **Push** your work back up to your fork
 5. Submit a **Pull request** so that we can review your changes

NOTE: Be sure to merge the latest from "upstream" before making a pull request!

[goversion-image]: https://img.shields.io/badge/Go-1.12+-00ADD8.svg
[goversion-url]: https://golang.org/
[godoc-image]: https://img.shields.io/badge/godoc-reference-00ADD8.svg
[godoc-url]: https://godoc.org/go.krak3n.codes/ocaws
[circle-image]: https://circleci.com/gh/krak3n/ocaws.svg?style=shield
[circle-url]: https://circleci.com/gh/krak3n/ocaws
[cover-image]: https://codecov.io/gh/krak3n/ocaws/branch/master/graph/badge.svg
[cover-url]: https://codecov.io/gh/krak3n/ocaws
[report-image]: https://goreportcard.com/badge/github.com/krak3n/ocaws
[report-url]: https://goreportcard.com/report/github.com/krak3n/ocaws
[jaeger]: assets/jaeger.png "Jaeger Tracing Example"
