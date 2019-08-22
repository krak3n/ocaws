package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/viper"
)

func main() {
	ctx := context.Background()

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	RegisterExporter()

	sns := NewSNS()
	sqs := NewSQS()

	if viper.GetBool("localstack") {
		viper.Set("sns.endpoint", "http://localhost:4575")
		viper.Set("sqs.endpoint", "http://localhost:4576")

		sns = NewSNS()
		sqs = NewSQS()

		if err := Localstack(sns, sqs); err != nil {
			log.Fatal("main: Localstack Error: ", err)
		}
	}

	errC := make(chan error, 1)
	sigC := make(chan os.Signal, 1)

	server := NewServer(sns)
	go func() {
		log.Println("main: Server Started:", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			errC <- err
		}
	}()

	defer server.Shutdown(ctx)

	client := NewClient(sqs)
	go func() {
		if err := client.Consume(ctx, DefaultHandler); err != nil {
			errC <- err
		}
	}()

	defer client.Stop()

	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case err := <-errC:
		log.Println("main: runtime error", err)
	case sig := <-sigC:
		log.Println("main: shutting down:", sig)
	}
}
