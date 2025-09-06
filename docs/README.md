# Cloud SDK Documentation

A multi-cloud SDK for Go, providing a unified API for AWS, GCP, Azure, etc. Inspired by Vercel AI SDK's developer experience.

## Features

- **Unified API**: Same interface across all cloud providers
- **Simple Initialization**: Easy provider setup and switching
- **Services**: Compute, Storage, Database operations
- **Extensible**: Add new providers with minimal changes

## Installation

```bash
go get github.com/VAIBHAVSING/Cloudsdk/go
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/VAIBHAVSING/Cloudsdk/go"
	"github.com/VAIBHAVSING/Cloudsdk/go/providers/aws"
)

func main() {
	ctx := context.Background()

	// Create AWS provider (Vercel AI SDK-style DX)
	provider, err := aws.Create(ctx, "us-east-1") // Clean, simple API
	// Alternative: aws.NewAWSProvider(ctx, "us-east-1") // Standard constructor
	if err != nil {
		log.Fatal(err)
	}

	// Initialize client
	client := cloudsdk.New(provider, &cloudsdk.Config{Region: "us-east-1"})

	// Create a VM
	vm, err := client.Compute().CreateVM(ctx, &services.VMConfig{
		Name:         "my-vm",
		ImageID:      "ami-12345678",
		InstanceType: "t2.micro",
		KeyName:      "my-key",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created VM: %s\n", vm.ID)

	// List VMs
	vms, err := client.Compute().ListVMs(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total VMs: %d\n", len(vms))

	// Create a bucket
	err = client.Storage().CreateBucket(ctx, &services.BucketConfig{
		Name:   "my-bucket",
		Region: "us-east-1",
	})
	if err != nil {
		log.Fatal(err)
	}

	// List buckets
	buckets, err := client.Storage().ListBuckets(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Buckets: %v\n", buckets)
}
```

## Supported Services

### Compute
- CreateVM
- ListVMs
- GetVM
- StartVM
- StopVM
- DeleteVM

### Storage
- CreateBucket
- ListBuckets
- DeleteBucket
- PutObject
- GetObject
- DeleteObject
- ListObjects

### Database
- CreateDB
- ListDBs
- GetDB
- DeleteDB

## Providers

### AWS
Currently implemented with full support for EC2, S3, and RDS.

### Future Providers
- GCP
- Azure
- Others

## Roadmap

- [ ] Declarative orchestration (JSON/YAML configurations)
- [ ] Additional providers
- [ ] CLI tool
- [ ] More comprehensive tests
- [ ] Documentation improvements

## Contributing

Contributions welcome! Please open issues or PRs on GitHub.

## License

See LICENSE file.
