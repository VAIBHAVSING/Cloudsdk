package main

import (
	"context"
	"fmt"
	"log"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	AWSProvider "github.com/VAIBHAVSING/Cloudsdk/go/providers/aws"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
)

func main() {
	fmt.Println("=== Cloud SDK AWS Example ===")

	// Create context
	ctx := context.Background()

	// Initialize AWS provider
	provider, err := AWSProvider.New("us-east-1")
	if err != nil {
		log.Fatalf("Failed to create AWS provider: %v", err)
	}

	// Optional: Test connection
	if err := provider.Connect(ctx); err != nil {
		log.Printf("AWS connection failed: %v", err)
		fmt.Println("Note: This is expected if AWS credentials are not configured")
	} else {
		fmt.Println("✓ Connected to AWS successfully")
	}

	// Create Cloud SDK client
	client := cloudsdk.NewFromProvider(provider)

	// Example 1: List VMs
	fmt.Println("\n--- Compute Service ---")
	vms, err := client.Compute().ListVMs(ctx)
	if err != nil {
		fmt.Printf("Error listing VMs: %v\n", err)
	} else {
		fmt.Printf("Found %d VMs\n", len(vms))
		for _, vm := range vms {
			fmt.Printf("  - %s (%s): %s\n", vm.Name, vm.ID, vm.State)
		}
	}

	// Example 2: List Buckets
	fmt.Println("\n--- Storage Service ---")
	buckets, err := client.Storage().ListBuckets(ctx)
	if err != nil {
		fmt.Printf("Error listing buckets: %v\n", err)
	} else {
		fmt.Printf("Found %d buckets\n", len(buckets))
		for _, bucket := range buckets {
			fmt.Printf("  - %s\n", bucket)
		}
	}

	// Example 3: Create a bucket (optional)
	bucketName := "cloudsdk-example-bucket-" + fmt.Sprintf("%d", 12345)
	fmt.Printf("\nTrying to create bucket: %s\n", bucketName)
	err = client.Storage().CreateBucket(ctx, &services.BucketConfig{
		Name:   bucketName,
	})
	if err != nil {
		fmt.Printf("Error creating bucket: %v\n", err)
	} else {
		fmt.Printf("✓ Bucket created successfully\n")
	}

	// Example 4: List Databases
	fmt.Println("\n--- Database Service ---")
	dbs, err := client.Database().ListDBs(ctx)
	if err != nil {
		fmt.Printf("Error listing databases: %v\n", err)
	} else {
		fmt.Printf("Found %d databases\n", len(dbs))
		for _, db := range dbs {
			fmt.Printf("  - %s (%s): %s\n", db.Name, db.Engine, db.Status)
		}
	}

	fmt.Println("\n=== Example completed ===")
}
