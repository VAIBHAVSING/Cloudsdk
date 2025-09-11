package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudsdk "github.com/VAIBHAVSING/Cloudsdk/go"
	awsprovider "github.com/VAIBHAVSING/Cloudsdk/go/providers/aws"
	"github.com/VAIBHAVSING/Cloudsdk/go/services"
)

func main() {
	// Create context
	ctx := context.Background()

	// Initialize AWS provider (two ways - similar to Vercel AI SDK DX)

	// Method 1: Standard constructor
	// provider, err := aws.NewAWSProvider(ctx, "us-east-1")

	// Method 2: Cleaner API similar to Vercel AI SDK
	provider, err := awsprovider.Create(ctx, "us-east-1")
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

	// Example 4: EC2 Instance Types
	fmt.Println("\n--- EC2 Instance Types ---")
	instanceTypes, err := client.Compute().InstanceTypes().List(ctx, &services.InstanceTypeFilter{
		VCpus: aws.Int32(2),
	})
	if err != nil {
		log.Printf("Error listing instance types: %v", err)
	} else {
		fmt.Printf("Found %d instance types with 2 vCPUs\n", len(instanceTypes))
		for i, it := range instanceTypes {
			if i >= 5 { // Limit output
				fmt.Println("... and more")
				break
			}
			fmt.Printf("- %s: %d vCPUs, %.1f GB RAM, %s network\n",
				it.InstanceType, it.VCpus, it.MemoryGB, it.NetworkPerformance)
		}
	}

	// Example 5: Placement Groups
	fmt.Println("\n--- Placement Groups ---")
	placementGroups, err := client.Compute().PlacementGroups().List(ctx)
	if err != nil {
		log.Printf("Error listing placement groups: %v", err)
	} else {
		fmt.Printf("Found %d placement groups\n", len(placementGroups))
		for _, pg := range placementGroups {
			fmt.Printf("- %s (%s): %s\n", pg.GroupName, pg.Strategy, pg.State)
		}
	}

	// Example 6: Spot Instances (Describe existing requests)
	fmt.Println("\n--- Spot Instance Requests ---")
	spotRequests, err := client.Compute().SpotInstances().Describe(ctx, nil)
	if err != nil {
		log.Printf("Error describing spot instance requests: %v", err)
	} else {
		fmt.Printf("Found %d spot instance requests\n", len(spotRequests))
		for _, req := range spotRequests {
			fmt.Printf("- Request %s: %s, Spot Price: %s\n",
				req.SpotInstanceRequestId, req.State, req.SpotPrice)
		}
	}

	fmt.Println("\n=== Example completed ===")
}
