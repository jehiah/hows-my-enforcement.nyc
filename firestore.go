package main

import (
	"context"

	"cloud.google.com/go/firestore"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func createClient(ctx context.Context) *firestore.Client {
	// Sets your Google Cloud Platform project ID.
	projectID := "hows-my-enforcement-nyc"

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Close client when done with
	// defer client.Close()
	return client
}

func IsAlreadyExists(err error) bool {
	return status.Code(err) == codes.AlreadyExists
}
func IsNotFound(err error) bool {
	return status.Code(err) == codes.NotFound
}
