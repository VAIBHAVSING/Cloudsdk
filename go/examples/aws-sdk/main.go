package main

import (
	"context"
	"fmt"
	"log"

	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/providers/aws"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
)

func main() {
	// Create context
	ctx := context.Background()

	// Initialize AWS provider (two ways - similar to Vercel AI SDK DX)

	// Method 1: Standard constructor
	// provider, err := aws.NewAWSProvider(ctx, "us-east-1")

	// Method 2: Cleaner API similar to Vercel AI SDK
	provider, err := aws.Create(ctx, "us-east-1")
	if err != nil {
		log.Fatalf("Failed to create AWS provider: %v", err)
	}

	// Create Cloud SDK client
	client := cloudsdk.New(provider, &cloudsdk.Config{Region: "us-east-1"})

	fmt.Println("=== Cloud SDK Example ===")

	// Example 1: Compute - List VMs
	fmt.Println("\n--- Computing Services ---")
	vms, err := client.Compute().ListVMs(ctx)
	if err != nil {
		log.Printf("Error listing VMs: %v", err)
	} else {
		fmt.Printf("Found %d VMs\n", len(vms))
		for _, vm := range vms {
			fmt.Printf("- VM ID: %s, State: %s\n", vm.ID, vm.State)
		}
	}

	// Example 2: Storage - Create Bucket
	fmt.Println("\n--- Storage Services ---")
	err = client.Storage().CreateBucket(ctx, &services.BucketConfig{
		Name:   "cloudsdk-example-bucket",
		Region: "us-east-1",
	})
	if err != nil {
		log.Printf("Error creating bucket: %v", err)
	} else {
		fmt.Println("Bucket created successfully")
	}

	// List buckets
	buckets, err := client.Storage().ListBuckets(ctx)
	if err != nil {
		log.Printf("Error listing buckets: %v", err)
	} else {
		fmt.Printf("Found %d buckets\n", len(buckets))
		for _, bucket := range buckets {
			fmt.Printf("- %s\n", bucket)
		}
	}

	// Example 3: Database - List DB instances
	fmt.Println("\n--- Database Services ---")
	dbs, err := client.Database().ListDBs(ctx)
	if err != nil {
		log.Printf("Error listing DBs: %v", err)
	} else {
		fmt.Printf("Found %d database instances\n", len(dbs))
		for _, db := range dbs {
			fmt.Printf("- DB ID: %s, Engine: %s, Status: %s\n", db.ID, db.Engine, db.Status)
		}
	}

	fmt.Println("\n=== Example completed ===")
}
