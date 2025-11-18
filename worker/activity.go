package main

import "context"

// SayHello is a simple activity that returns a greeting message.
func SayHello(ctx context.Context, name string) (string, error) {
	return "Hello, " + name + " from Temporal Worker on AWS!", nil
}

