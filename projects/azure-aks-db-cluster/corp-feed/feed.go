package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	clientOptions := azcosmos.ClientOptions{
		EnableContentResponseOnWrite: true,
	}

	client, err := azcosmos.NewClient("<azure-cosmos-db-nosql-account-endpoint>", credential, &clientOptions)
	if err != nil {
		return err
	}

	client.CreateDatabase(ctx, azcosmos.DatabaseProperties{}, &azcosmos.CreateDatabaseOptions{
		ThroughputProperties: &azcosmos.ThroughputProperties{

		},
	})

	db, err := client.NewDatabase("10")
	resp, err := db.Read(ctx, &azcosmos.ReadDatabaseOptions{})

	return nil
}

func main() {
	ctx := context.Background()

	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
