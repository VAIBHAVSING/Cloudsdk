# AWS Cloud SDK Example

A simple example demonstrating how to use the Cloud SDK with AWS.

## What This Example Shows

- How to create an AWS provider
- How to use the unified Cloud SDK client
- Basic operations across Compute, Storage, and Database services
- Simple error handling

## Prerequisites

You need AWS credentials configured. Choose one of these methods:

### Option 1: Environment Variables
```bash
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
```

### Option 2: AWS CLI Profile
```bash
aws configure
# or
aws configure --profile myprofile
```

### Option 3: IAM Role (when running on AWS)
No configuration needed - automatic when running on EC2, ECS, or Lambda.

## Running the Example

```bash
cd go/examples/aws-sdk
go run main.go
```

## Expected Output

The example will:
1. Create an AWS provider for the `us-east-1` region
2. Test the connection to AWS
3. List your existing VMs (EC2 instances)
4. List your existing S3 buckets
5. Try to create a new S3 bucket
6. List your existing RDS databases

## Code Walkthrough

```go
// 1. Create AWS provider
provider, err := AWSProvider.New("us-east-1")

// 2. Create unified client
client := cloudsdk.NewFromProvider(provider)

// 3. Use any service with the same client
vms, err := client.Compute().ListVMs(ctx)
buckets, err := client.Storage().ListBuckets(ctx)
dbs, err := client.Database().ListDBs(ctx)
```

## Error Handling

The example shows basic error handling. If you don't have AWS credentials configured, you'll see helpful error messages explaining how to set them up.

## Next Steps

- Check out the [service availability example](../service-availability/) for more advanced features
- See the [provider documentation](../../providers/aws/) for all available options
- Try different AWS regions by changing the region parameter