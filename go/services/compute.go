package services

import "context"

// VMConfig represents the configuration for creating a virtual machine.
// All fields are validated before creating the VM to ensure proper configuration.
// Supports JSON/YAML serialization for external configuration files.
//
// Validation Rules:
//   - Name: 1-255 characters, alphanumeric and hyphens only, cannot start/end with hyphen
//   - ImageID: Must exist in target region and be accessible to your account
//   - InstanceType: Must be available in target region and supported by chosen image
//   - KeyName: Must exist in target region (if specified)
//   - SecurityGroups: All groups must exist and be accessible
//   - UserData: Maximum 16KB, automatically base64 encoded when needed
//   - Tags: Maximum 50 tags, each key/value max 255 characters
//
// Provider-Specific Behaviors:
//   - AWS: SecurityGroups can be IDs (sg-xxx) or names, UserData auto-base64 encoded
//   - GCP: Uses network tags instead of security groups, UserData as startup-script
//   - Azure: Uses Network Security Groups, UserData as custom-data
//
// Example:
//
//	config := &VMConfig{
//	    Name:         "web-server-01",
//	    ImageID:      "ami-12345678",  // Amazon Linux 2 AMI
//	    InstanceType: "t2.micro",     // Free tier eligible
//	    KeyName:      "my-keypair",   // Must exist in the region
//	    SecurityGroups: []string{"sg-web", "sg-ssh"},
//	    UserData:     "#!/bin/bash\nyum update -y",
//	    Tags: map[string]string{
//	        "Environment": "production",
//	        "Team":        "backend",
//	        "Project":     "web-app",
//	    },
//	    SubnetID:     "subnet-12345",
//	    AssignPublicIP: true,
//	}
type VMConfig struct {
	// Name is the display name for the virtual machine.
	// Validation: 1-255 characters, alphanumeric and hyphens only
	// Cannot start or end with hyphen, must be unique within account/region
	// Used for identification and billing purposes
	//
	// Examples: "web-server-01", "database-primary", "worker-node-3"
	Name string `json:"name" yaml:"name" validate:"required,min=1,max=255,hostname"`

	// ImageID is the identifier of the machine image to use.
	// Must exist in target region and be accessible to your account
	// Determines the operating system and pre-installed software
	//
	// Format varies by provider:
	//   - AWS: ami-xxxxxxxx (e.g., "ami-12345678")
	//   - GCP: projects/PROJECT/global/images/IMAGE or family/FAMILY
	//   - Azure: publisher:offer:sku:version (e.g., "Canonical:UbuntuServer:18.04-LTS:latest")
	//
	// Common Images:
	//   - AWS: ami-0abcdef1234567890 (Amazon Linux 2), ami-0987654321fedcba0 (Ubuntu 20.04)
	//   - GCP: "projects/ubuntu-os-cloud/global/images/family/ubuntu-2004-lts"
	//   - Azure: "Canonical:0001-com-ubuntu-server-focal:20_04-lts-gen2:latest"
	ImageID string `json:"image_id" yaml:"image_id" validate:"required"`

	// InstanceType specifies the hardware configuration for the VM.
	// Must be available in target region and compatible with chosen image
	// Determines CPU, memory, storage, and network performance
	//
	// Common types by provider:
	//   AWS:
	//     - t2.micro (1 vCPU, 1GB RAM) - Free tier eligible
	//     - t2.small (1 vCPU, 2GB RAM) - General purpose
	//     - m5.large (2 vCPU, 8GB RAM) - Balanced compute
	//     - c5.xlarge (4 vCPU, 8GB RAM) - Compute optimized
	//     - r5.large (2 vCPU, 16GB RAM) - Memory optimized
	//   GCP:
	//     - e2-micro (0.25-2 vCPU, 1GB RAM) - Shared core
	//     - e2-small (0.5-2 vCPU, 2GB RAM) - Shared core
	//     - n1-standard-1 (1 vCPU, 3.75GB RAM) - Standard
	//     - c2-standard-4 (4 vCPU, 16GB RAM) - Compute optimized
	//   Azure:
	//     - Standard_B1s (1 vCPU, 1GB RAM) - Burstable
	//     - Standard_B2s (2 vCPU, 4GB RAM) - Burstable
	//     - Standard_D2s_v3 (2 vCPU, 8GB RAM) - General purpose
	InstanceType string `json:"instance_type" yaml:"instance_type" validate:"required"`

	// KeyName is the name of the SSH key pair for secure access.
	// Must exist in target region before creating the VM
	// Leave empty if using other authentication methods (passwords, certificates)
	//
	// Key Management:
	//   - AWS: Key pairs are region-specific, create via EC2 console or CLI
	//   - GCP: Uses project-wide SSH keys or instance-specific keys
	//   - Azure: Can use SSH keys or passwords for authentication
	//
	// Security Best Practices:
	//   - Use strong SSH keys (RSA 2048+ bits or Ed25519)
	//   - Rotate keys regularly
	//   - Use different keys for different environments
	//   - Store private keys securely
	KeyName string `json:"key_name,omitempty" yaml:"key_name,omitempty"`

	// SecurityGroups define the firewall rules for the VM.
	// Each security group controls inbound and outbound traffic
	// At least one security group is required for most providers
	//
	// Provider Differences:
	//   - AWS: Can specify by ID (sg-xxx) or name, supports multiple groups
	//   - GCP: Uses network tags and firewall rules instead
	//   - Azure: Uses Network Security Groups (NSGs)
	//
	// Common Patterns:
	//   - ["sg-web"]: HTTP/HTTPS access (ports 80, 443)
	//   - ["sg-ssh"]: SSH access (port 22)
	//   - ["sg-database"]: Database access (port 3306, 5432, etc.)
	//   - ["sg-web", "sg-ssh"]: Web server with SSH access
	//
	// Security Best Practices:
	//   - Use principle of least privilege
	//   - Create specific security groups for different roles
	//   - Avoid using default security groups in production
	//   - Regularly audit security group rules
	SecurityGroups []string `json:"security_groups,omitempty" yaml:"security_groups,omitempty"`

	// UserData contains initialization scripts that run when the VM starts.
	// Executed with root/administrator privileges during first boot
	// Maximum size: 16KB for most providers (automatically enforced)
	//
	// Format Requirements:
	//   - Linux: Shell script starting with #!/bin/bash or cloud-init YAML
	//   - Windows: PowerShell script or batch commands
	//   - Automatically base64 encoded when required by provider
	//
	// Common Use Cases:
	//   - Install and configure software packages
	//   - Download application code from repositories
	//   - Configure system settings and services
	//   - Set up monitoring and logging agents
	//   - Join domain or cluster
	//
	// Examples:
	//   Linux: "#!/bin/bash\nyum update -y\nyum install -y httpd\nsystemctl start httpd"
	//   Windows: "powershell.exe Install-WindowsFeature -Name Web-Server"
	//   Cloud-init: "#cloud-config\npackages:\n  - nginx\n  - git"
	//
	// Best Practices:
	//   - Keep scripts idempotent (safe to run multiple times)
	//   - Log script execution for debugging
	//   - Use configuration management tools for complex setups
	//   - Test scripts thoroughly before production use
	UserData string `json:"user_data,omitempty" yaml:"user_data,omitempty" validate:"max=16384"`

	// Tags are key-value pairs for organizing and managing resources.
	// Used for billing, automation, access control, and resource organization
	// Maximum 50 tags per resource, each key/value max 255 characters
	//
	// Common Tag Patterns:
	//   - Environment: "production", "staging", "development"
	//   - Team: "backend", "frontend", "devops"
	//   - Project: "web-app", "mobile-api", "data-pipeline"
	//   - Owner: "john.doe@company.com"
	//   - CostCenter: "engineering", "marketing"
	//   - Backup: "daily", "weekly", "none"
	//   - Monitoring: "enabled", "disabled"
	//
	// Provider-Specific Notes:
	//   - AWS: Tags are case-sensitive, support resource-based policies
	//   - GCP: Called "labels", lowercase keys/values only
	//   - Azure: Called "tags", case-insensitive
	//
	// Best Practices:
	//   - Use consistent naming conventions across organization
	//   - Include mandatory tags (Environment, Owner, Project)
	//   - Use tags for cost allocation and chargeback
	//   - Automate tag compliance with policies
	Tags map[string]string `json:"tags,omitempty" yaml:"tags,omitempty" validate:"max=50,dive,keys,max=255,endkeys,max=255"`

	// SubnetID specifies the subnet where the VM should be launched.
	// Must exist in the target region and be accessible to your account
	// Determines the network segment and availability zone for the VM
	//
	// Network Planning:
	//   - Public subnets: Have internet gateway, get public IPs
	//   - Private subnets: No direct internet access, use NAT gateway
	//   - Choose subnet in appropriate availability zone
	//   - Ensure subnet has sufficient IP addresses available
	//
	// Format by Provider:
	//   - AWS: subnet-xxxxxxxx (e.g., "subnet-12345678")
	//   - GCP: projects/PROJECT/regions/REGION/subnetworks/SUBNET
	//   - Azure: /subscriptions/.../resourceGroups/.../providers/Microsoft.Network/virtualNetworks/.../subnets/...
	//
	// Leave empty to use default subnet for the region
	SubnetID string `json:"subnet_id,omitempty" yaml:"subnet_id,omitempty"`

	// AssignPublicIP determines whether the VM gets a public IP address.
	// Public IPs allow direct internet access but increase security exposure
	// Default behavior varies by provider and subnet configuration
	//
	// Considerations:
	//   - Public IPs: Direct internet access, higher security risk, additional cost
	//   - Private IPs: More secure, requires NAT gateway or VPN for internet access
	//   - Load balancers: Can provide public access without public IPs on instances
	//
	// Provider Defaults:
	//   - AWS: Depends on subnet's "auto-assign public IP" setting
	//   - GCP: No public IP by default
	//   - Azure: Depends on subnet and VM configuration
	//
	// Security Best Practices:
	//   - Use private IPs for internal services
	//   - Use load balancers for public-facing applications
	//   - Implement proper security groups/firewall rules
	//   - Consider using bastion hosts for SSH access
	AssignPublicIP *bool `json:"assign_public_ip,omitempty" yaml:"assign_public_ip,omitempty"`

	// PlacementGroup specifies a placement group for the VM.
	// Placement groups influence how instances are placed on underlying hardware
	// Used to optimize for network performance, fault tolerance, or both
	//
	// Placement Strategies:
	//   - Cluster: Low-latency networking within single AZ (best performance)
	//   - Partition: Spread across logical partitions (reduce correlated failures)
	//   - Spread: Spread across distinct hardware (maximum fault tolerance)
	//
	// Use Cases:
	//   - HPC applications: Use cluster placement for low latency
	//   - Distributed systems: Use partition placement for fault tolerance
	//   - Critical applications: Use spread placement for maximum availability
	//
	// Leave empty if placement optimization is not required
	PlacementGroup string `json:"placement_group,omitempty" yaml:"placement_group,omitempty"`

	// IamInstanceProfile specifies the IAM role for the VM.
	// Provides AWS credentials and permissions to applications running on the VM
	// More secure than embedding access keys in code or configuration
	//
	// Benefits:
	//   - No need to store AWS credentials on the instance
	//   - Automatic credential rotation
	//   - Fine-grained permissions control
	//   - Audit trail of API calls
	//
	// Format: IAM instance profile name or ARN
	// Examples: "EC2-S3-Access-Role", "arn:aws:iam::123456789012:instance-profile/MyRole"
	//
	// Best Practices:
	//   - Use least privilege principle
	//   - Create specific roles for different application types
	//   - Regularly review and audit role permissions
	//   - Use AWS managed policies when possible
	IamInstanceProfile string `json:"iam_instance_profile,omitempty" yaml:"iam_instance_profile,omitempty"`

	// Monitoring enables detailed monitoring for the VM.
	// Provides additional metrics and shorter metric intervals
	// May incur additional costs depending on provider
	//
	// Benefits:
	//   - More frequent metric collection (1-minute vs 5-minute intervals)
	//   - Additional system and application metrics
	//   - Better visibility into performance and health
	//   - Faster detection of issues
	//
	// Considerations:
	//   - Additional cost for detailed monitoring
	//   - More data storage and network usage
	//   - May not be necessary for all workloads
	//
	// Default: false (basic monitoring only)
	Monitoring *bool `json:"monitoring,omitempty" yaml:"monitoring,omitempty"`

	// EbsOptimized enables EBS optimization for better storage performance.
	// Provides dedicated bandwidth for EBS volumes separate from network traffic
	// Not all instance types support EBS optimization
	//
	// Benefits:
	//   - Better EBS volume performance
	//   - Reduced network contention
	//   - More consistent storage throughput
	//   - Better performance for I/O intensive applications
	//
	// Considerations:
	//   - Additional cost for some instance types
	//   - Not supported on all instance types
	//   - May not be necessary for low I/O workloads
	//
	// Default: false (not EBS optimized)
	EbsOptimized *bool `json:"ebs_optimized,omitempty" yaml:"ebs_optimized,omitempty"`
}

// VM represents a virtual machine instance with its current state and network information.
// This struct provides a unified view of VMs across different cloud providers.
type VM struct {
	// ID is the unique identifier assigned by the cloud provider.
	// Format varies by provider:
	//   - AWS: i-1234567890abcdef0
	//   - GCP: 1234567890123456789
	//   - Azure: /subscriptions/.../resourceGroups/.../providers/Microsoft.Compute/virtualMachines/vm-name
	ID string

	// Name is the human-readable name of the virtual machine.
	// This is the name specified during creation and may not be unique across the account.
	Name string

	// State represents the current operational state of the VM.
	// Common states: "pending", "running", "stopping", "stopped", "terminating", "terminated"
	// State transitions may take several minutes depending on the provider and instance type.
	State string

	// PublicIP is the internet-accessible IP address of the VM.
	// May be empty if the VM is in a private subnet or doesn't have a public IP assigned.
	// Format: IPv4 address (e.g., "203.0.113.1")
	PublicIP string

	// PrivateIP is the internal network IP address of the VM.
	// Used for communication within the virtual private cloud (VPC).
	// Format: IPv4 address (e.g., "10.0.1.100")
	PrivateIP string

	// LaunchTime indicates when the VM was created.
	// Format: RFC3339 timestamp (e.g., "2023-01-15T10:30:00Z")
	LaunchTime string
}

// InstanceTypeFilter represents filters for querying available instance types.
// All filter fields are optional - use nil/empty values to skip filtering on that attribute.
//
// Example:
//
//	filter := &InstanceTypeFilter{
//	    VCpus:    aws.Int32(2),           // Only 2-CPU instances
//	    MemoryGB: aws.Float64(4.0),       // Only instances with 4GB RAM
//	    NetworkPerf: aws.String("High"),  // Only high-performance networking
//	}
type InstanceTypeFilter struct {
	// VCpus filters by the number of virtual CPUs.
	// Use nil to include all CPU configurations.
	VCpus *int32

	// MemoryGB filters by the amount of memory in gigabytes.
	// Use nil to include all memory configurations.
	MemoryGB *float64

	// StorageGB filters by the amount of instance storage in gigabytes.
	// Use nil to include all storage configurations.
	// Note: This refers to ephemeral storage, not EBS volumes.
	StorageGB *int32

	// NetworkPerf filters by network performance level.
	// Common values: "Low", "Moderate", "High", "Up to 10 Gigabit", "25 Gigabit"
	// Use nil to include all network performance levels.
	NetworkPerf *string

	// InstanceTypes filters by specific instance type names.
	// Use empty slice to include all instance types.
	// Example: []string{"t2.micro", "t2.small", "m5.large"}
	InstanceTypes []string
}

// InstanceType represents the specifications of a virtual machine instance type.
// This provides detailed hardware information to help choose the right instance for your workload.
type InstanceType struct {
	// InstanceType is the name/identifier of the instance type.
	// Examples: "t2.micro", "m5.large", "c5.xlarge"
	InstanceType string

	// VCpus is the number of virtual CPU cores available.
	// Note: Some instance types use shared CPU (burstable performance).
	VCpus int32

	// MemoryGB is the amount of RAM in gigabytes.
	// This is the total memory available to the operating system and applications.
	MemoryGB float64

	// StorageGB is the amount of ephemeral (instance) storage in gigabytes.
	// This is temporary storage that is lost when the instance stops.
	// Use 0 for EBS-only instance types.
	StorageGB int32

	// NetworkPerformance describes the network bandwidth capability.
	// Examples: "Low", "Moderate", "High", "Up to 10 Gigabit", "25 Gigabit"
	NetworkPerformance string

	// CurrentGeneration indicates if this is a current-generation instance type.
	// Previous generation instances may have lower performance or higher costs.
	// Prefer current generation instances for new deployments.
	CurrentGeneration bool
}

// InstanceTypesService provides operations for querying available virtual machine instance types.
// Use this service to discover what hardware configurations are available in your region.
type InstanceTypesService interface {
	// List retrieves available instance types, optionally filtered by specifications.
	// Returns all available instance types in the current region if no filter is provided.
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials or expired tokens
	//   - ErrAuthorization: Insufficient permissions to describe instance types
	//   - ErrInvalidConfig: Invalid filter parameters
	//
	// Example:
	//   // Get all instance types
	//   types, err := compute.InstanceTypes().List(ctx, nil)
	//
	//   // Get only 2-CPU instances with at least 4GB RAM
	//   filter := &InstanceTypeFilter{
	//       VCpus:    aws.Int32(2),
	//       MemoryGB: aws.Float64(4.0),
	//   }
	//   types, err := compute.InstanceTypes().List(ctx, filter)
	//
	//   for _, instanceType := range types {
	//       fmt.Printf("%s: %d vCPUs, %.1f GB RAM\n",
	//           instanceType.InstanceType, instanceType.VCpus, instanceType.MemoryGB)
	//   }
	List(ctx context.Context, filter *InstanceTypeFilter) ([]*InstanceType, error)
}

// PlacementGroupConfig represents configuration for creating a placement group.
// Placement groups influence how instances are placed on underlying hardware
// to optimize for network performance, fault tolerance, or both.
//
// Example:
//
//	config := &PlacementGroupConfig{
//	    GroupName: "high-perf-cluster",
//	    Strategy:  "cluster",  // For low-latency networking
//	}
type PlacementGroupConfig struct {
	// GroupName is the unique name for the placement group.
	// Must be unique within your account and region.
	// Length: 1-255 characters, alphanumeric and hyphens only.
	GroupName string

	// Strategy determines how instances are placed on underlying hardware.
	// Valid values:
	//   - "cluster": Groups instances in a low-latency group in a single AZ (best network performance)
	//   - "partition": Spreads instances across logical partitions (reduces correlated failures)
	//   - "spread": Spreads instances across distinct underlying hardware (maximum fault tolerance)
	Strategy string
}

// PlacementGroup represents a placement group that controls instance placement on hardware.
// Placement groups help optimize network performance and fault tolerance for your workloads.
type PlacementGroup struct {
	// GroupName is the user-specified name of the placement group.
	GroupName string

	// GroupId is the unique identifier assigned by the cloud provider.
	GroupId string

	// Strategy indicates how instances are placed within this group.
	// Values: "cluster", "partition", "spread"
	Strategy string

	// State represents the current state of the placement group.
	// Common states: "pending", "available", "deleting", "deleted"
	State string

	// GroupArn is the Amazon Resource Name (ARN) for AWS placement groups.
	// Format: arn:aws:ec2:region:account:placement-group/group-name
	// May be empty for other cloud providers.
	GroupArn string
}

// PlacementGroupsService provides operations for managing placement groups.
// Placement groups control how instances are positioned on underlying hardware
// to optimize for performance, availability, or both.
type PlacementGroupsService interface {
	// Create creates a new placement group with the specified configuration.
	// The placement group must be created before launching instances into it.
	//
	// Common errors:
	//   - ErrResourceConflict: Placement group name already exists
	//   - ErrInvalidConfig: Invalid strategy or group name
	//   - ErrAuthorization: Insufficient permissions to create placement groups
	//
	// Example:
	//   config := &PlacementGroupConfig{
	//       GroupName: "web-cluster",
	//       Strategy:  "cluster",
	//   }
	//   group, err := compute.PlacementGroups().Create(ctx, config)
	//   if err != nil {
	//       log.Fatalf("Failed to create placement group: %v", err)
	//   }
	//   fmt.Printf("Created placement group: %s\n", group.GroupName)
	Create(ctx context.Context, config *PlacementGroupConfig) (*PlacementGroup, error)

	// Delete removes a placement group permanently.
	// The placement group must be empty (no instances) before it can be deleted.
	// This operation cannot be undone.
	//
	// Common errors:
	//   - ErrResourceNotFound: Placement group doesn't exist
	//   - ErrResourceConflict: Placement group still contains instances
	//   - ErrAuthorization: Insufficient permissions to delete placement groups
	//
	// Example:
	//   err := compute.PlacementGroups().Delete(ctx, "old-cluster")
	//   if err != nil {
	//       log.Printf("Failed to delete placement group: %v", err)
	//   }
	Delete(ctx context.Context, groupName string) error

	// List retrieves all placement groups in the current region.
	// Returns an empty slice if no placement groups exist.
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials
	//   - ErrAuthorization: Insufficient permissions to list placement groups
	//
	// Example:
	//   groups, err := compute.PlacementGroups().List(ctx)
	//   if err != nil {
	//       log.Fatalf("Failed to list placement groups: %v", err)
	//   }
	//   for _, group := range groups {
	//       fmt.Printf("Group: %s, Strategy: %s, State: %s\n",
	//           group.GroupName, group.Strategy, group.State)
	//   }
	List(ctx context.Context) ([]*PlacementGroup, error)
}

// SpotInstanceConfig represents configuration for requesting spot instances.
// Spot instances offer significant cost savings but can be interrupted when demand increases.
// Best for fault-tolerant, flexible workloads that can handle interruptions.
//
// Example:
//
//	config := &SpotInstanceConfig{
//	    InstanceType:     "m5.large",
//	    ImageID:          "ami-12345678",
//	    SpotPrice:        aws.String("0.05"),  // Maximum price per hour
//	    AvailabilityZone: aws.String("us-east-1a"),
//	    LaunchSpecification: &SpotLaunchSpec{
//	        ImageID:      "ami-12345678",
//	        InstanceType: "m5.large",
//	        KeyName:      "my-key",
//	        SecurityGroups: []string{"sg-12345"},
//	    },
//	}
type SpotInstanceConfig struct {
	// InstanceType specifies the type of instance to request.
	// Must be available as a spot instance in the target region.
	InstanceType string

	// ImageID is the machine image to use for the spot instance.
	// Format: ami-xxxxxxxx for AWS
	ImageID string

	// SpotPrice is the maximum price you're willing to pay per hour.
	// Use nil to accept the current spot price.
	// Format: decimal string (e.g., "0.05" for $0.05/hour)
	// Tip: Check current spot prices to set competitive bids.
	SpotPrice *string

	// AvailabilityZone specifies where to launch the spot instance.
	// Use nil to let the provider choose the best zone.
	// Specific zones may have better pricing or availability.
	AvailabilityZone *string

	// LaunchSpecification defines the instance configuration.
	// This is similar to regular instance configuration but for spot requests.
	LaunchSpecification *SpotLaunchSpec
}

// SpotLaunchSpec represents the launch specification for spot instances.
// This defines how the spot instance should be configured when launched.
type SpotLaunchSpec struct {
	// ImageID is the machine image identifier.
	// Must be available in the target region and compatible with the instance type.
	ImageID string

	// InstanceType specifies the hardware configuration.
	// Must match the instance type in the parent SpotInstanceConfig.
	InstanceType string

	// KeyName is the SSH key pair for secure access.
	// The key pair must exist in the target region.
	KeyName string

	// SecurityGroups define firewall rules for the spot instance.
	// At least one security group is typically required.
	SecurityGroups []string

	// UserData contains initialization scripts for the spot instance.
	// Runs when the instance starts, useful for software installation and configuration.
	UserData string
}

// SpotInstanceRequest represents a spot instance request and its current status.
// Spot requests go through several states before an instance is launched.
type SpotInstanceRequest struct {
	// SpotInstanceRequestId is the unique identifier for this spot request.
	SpotInstanceRequestId string

	// InstanceId is the ID of the launched instance (if successful).
	// Empty if the request hasn't been fulfilled yet.
	InstanceId string

	// State represents the current state of the spot request.
	// Common states: "open", "active", "closed", "cancelled", "failed"
	State string

	// Status provides additional details about the request state.
	// Examples: "pending-evaluation", "pending-fulfillment", "fulfilled", "instance-terminated-by-price"
	Status string

	// SpotPrice is the current spot price for this instance type.
	// Format: decimal string (e.g., "0.0464")
	SpotPrice string

	// LaunchSpecification contains the instance configuration.
	// This is the configuration that will be used when the spot request is fulfilled.
	LaunchSpecification *SpotLaunchSpec

	// CreateTime indicates when the spot request was created.
	// Format: RFC3339 timestamp
	CreateTime string
}

// SpotInstancesService provides operations for managing spot instances.
// Spot instances offer significant cost savings (up to 90%) but can be interrupted
// when capacity is needed for on-demand instances.
type SpotInstancesService interface {
	// Request creates a new spot instance request with the specified configuration.
	// The request may take time to be fulfilled depending on capacity and pricing.
	//
	// Common errors:
	//   - ErrInvalidConfig: Invalid instance type, image ID, or spot price
	//   - ErrAuthorization: Insufficient permissions or spot instance limits exceeded
	//   - ErrResourceNotFound: Image ID or security groups don't exist
	//
	// Example:
	//   config := &SpotInstanceConfig{
	//       InstanceType: "m5.large",
	//       ImageID:      "ami-12345678",
	//       SpotPrice:    aws.String("0.10"),  // Max $0.10/hour
	//       LaunchSpecification: &SpotLaunchSpec{
	//           ImageID:        "ami-12345678",
	//           InstanceType:   "m5.large",
	//           KeyName:        "my-keypair",
	//           SecurityGroups: []string{"sg-web"},
	//       },
	//   }
	//   request, err := compute.SpotInstances().Request(ctx, config)
	//   if err != nil {
	//       log.Fatalf("Failed to request spot instance: %v", err)
	//   }
	//   fmt.Printf("Spot request created: %s\n", request.SpotInstanceRequestId)
	Request(ctx context.Context, config *SpotInstanceConfig) (*SpotInstanceRequest, error)

	// Describe retrieves the status of one or more spot instance requests.
	// Use empty slice to get all spot requests in the region.
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials
	//   - ErrAuthorization: Insufficient permissions to describe spot requests
	//   - ErrResourceNotFound: One or more request IDs don't exist
	//
	// Example:
	//   // Get specific requests
	//   requests, err := compute.SpotInstances().Describe(ctx, []string{"sir-12345", "sir-67890"})
	//
	//   // Get all requests
	//   allRequests, err := compute.SpotInstances().Describe(ctx, []string{})
	//
	//   for _, req := range requests {
	//       fmt.Printf("Request %s: %s (%s)\n", req.SpotInstanceRequestId, req.State, req.Status)
	//       if req.InstanceId != "" {
	//           fmt.Printf("  Instance: %s\n", req.InstanceId)
	//       }
	//   }
	Describe(ctx context.Context, requestIds []string) ([]*SpotInstanceRequest, error)

	// Cancel cancels a spot instance request.
	// If the request has already launched an instance, the instance will continue running
	// but the spot request will be cancelled to prevent replacement instances.
	//
	// Common errors:
	//   - ErrResourceNotFound: Spot request doesn't exist
	//   - ErrAuthorization: Insufficient permissions to cancel spot requests
	//   - ErrInvalidConfig: Request is already cancelled or in a non-cancellable state
	//
	// Example:
	//   err := compute.SpotInstances().Cancel(ctx, "sir-12345678")
	//   if err != nil {
	//       log.Printf("Failed to cancel spot request: %v", err)
	//   } else {
	//       fmt.Println("Spot request cancelled successfully")
	//   }
	Cancel(ctx context.Context, requestId string) error
}

// Compute provides virtual machine management operations across cloud providers.
// All methods return structured errors with helpful context and suggestions for troubleshooting.
// This interface abstracts the differences between AWS EC2, Google Compute Engine, Azure VMs, etc.
//
// Authentication and Authorization:
// All operations require valid cloud provider credentials with appropriate permissions.
// Ensure your credentials have compute instance management permissions before using this service.
//
// Rate Limiting:
// Cloud providers impose rate limits on API calls. The SDK automatically retries
// transient failures with exponential backoff, but you may still encounter rate limit errors
// during high-frequency operations.
//
// Regional Resources:
// Virtual machines are regional resources. Ensure you're operating in the correct region
// and that the specified resources (images, security groups, key pairs) exist in that region.
type Compute interface {
	// CreateVM creates a new virtual machine with the specified configuration.
	// The VM will be launched asynchronously - use GetVM to check the status.
	// Returns the VM details immediately, but the VM may still be in "pending" state.
	//
	// Parameter Validation:
	//   - Name: 1-255 characters, alphanumeric and hyphens only
	//   - ImageID: Must exist in the target region and be accessible to your account
	//   - InstanceType: Must be available in the target region
	//   - KeyName: Must exist in the target region (if specified)
	//   - SecurityGroups: All groups must exist in the target region
	//   - UserData: Maximum 16KB, will be base64 encoded automatically
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials or expired tokens
	//   - ErrAuthorization: Insufficient permissions or instance limits exceeded
	//   - ErrInvalidConfig: Invalid VM configuration (missing required fields, invalid values)
	//   - ErrResourceNotFound: Image ID, security groups, or key pair don't exist
	//   - ErrResourceConflict: VM name already exists (provider-dependent)
	//   - ErrRateLimit: Too many requests, retry with exponential backoff
	//
	// Example:
	//   config := &VMConfig{
	//       Name:         "web-server-01",
	//       ImageID:      "ami-12345678",        // Amazon Linux 2
	//       InstanceType: "t2.micro",           // Free tier eligible
	//       KeyName:      "my-keypair",         // For SSH access
	//       SecurityGroups: []string{"sg-web"}, // HTTP/HTTPS access
	//       UserData:     "#!/bin/bash\nyum update -y\nyum install -y httpd\nsystemctl start httpd",
	//   }
	//
	//   vm, err := compute.CreateVM(ctx, config)
	//   if err != nil {
	//       log.Fatalf("Failed to create VM: %v", err)
	//   }
	//
	//   fmt.Printf("VM created: %s (ID: %s, State: %s)\n", vm.Name, vm.ID, vm.State)
	//
	//   // Wait for VM to be running
	//   for vm.State == "pending" {
	//       time.Sleep(10 * time.Second)
	//       vm, err = compute.GetVM(ctx, vm.ID)
	//       if err != nil {
	//           log.Fatalf("Failed to check VM status: %v", err)
	//       }
	//   }
	//
	//   if vm.State == "running" {
	//       fmt.Printf("VM is ready! Public IP: %s\n", vm.PublicIP)
	//   }
	CreateVM(ctx context.Context, config *VMConfig) (*VM, error)

	// ListVMs returns all virtual machines in the current region.
	// Returns an empty slice if no VMs exist. Results are not paginated - all VMs are returned.
	// For accounts with many VMs, consider using provider-specific filtering if available.
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials or expired tokens
	//   - ErrAuthorization: Insufficient permissions to list instances
	//   - ErrRateLimit: Too many requests, retry with exponential backoff
	//
	// Example:
	//   vms, err := compute.ListVMs(ctx)
	//   if err != nil {
	//       log.Fatalf("Failed to list VMs: %v", err)
	//   }
	//
	//   fmt.Printf("Found %d VMs:\n", len(vms))
	//   for _, vm := range vms {
	//       fmt.Printf("  %s (%s): %s - %s\n", vm.Name, vm.ID, vm.State, vm.PublicIP)
	//   }
	//
	//   // Filter running VMs
	//   runningVMs := make([]*VM, 0)
	//   for _, vm := range vms {
	//       if vm.State == "running" {
	//           runningVMs = append(runningVMs, vm)
	//       }
	//   }
	//   fmt.Printf("%d VMs are currently running\n", len(runningVMs))
	ListVMs(ctx context.Context) ([]*VM, error)

	// GetVM retrieves detailed information about a specific virtual machine.
	// Returns the current state, network information, and metadata for the VM.
	// Use this method to check VM status after creation or state changes.
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials or expired tokens
	//   - ErrAuthorization: Insufficient permissions or VM not owned by your account
	//   - ErrResourceNotFound: VM with the specified ID doesn't exist
	//   - ErrRateLimit: Too many requests, retry with exponential backoff
	//
	// Example:
	//   vm, err := compute.GetVM(ctx, "i-1234567890abcdef0")
	//   if err != nil {
	//       if errors.Is(err, ErrResourceNotFound) {
	//           fmt.Println("VM not found - it may have been terminated")
	//           return
	//       }
	//       log.Fatalf("Failed to get VM details: %v", err)
	//   }
	//
	//   fmt.Printf("VM Details:\n")
	//   fmt.Printf("  Name: %s\n", vm.Name)
	//   fmt.Printf("  State: %s\n", vm.State)
	//   fmt.Printf("  Public IP: %s\n", vm.PublicIP)
	//   fmt.Printf("  Private IP: %s\n", vm.PrivateIP)
	//   fmt.Printf("  Launch Time: %s\n", vm.LaunchTime)
	//
	//   // Check if VM is accessible
	//   if vm.State == "running" && vm.PublicIP != "" {
	//       fmt.Printf("VM is accessible at: ssh -i ~/.ssh/my-key.pem ec2-user@%s\n", vm.PublicIP)
	//   }
	GetVM(ctx context.Context, id string) (*VM, error)

	// StartVM starts a stopped virtual machine.
	// The VM must be in "stopped" state. Starting may take several minutes depending on the instance type.
	// Use GetVM to monitor the state transition from "pending" to "running".
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials or expired tokens
	//   - ErrAuthorization: Insufficient permissions or VM not owned by your account
	//   - ErrResourceNotFound: VM with the specified ID doesn't exist
	//   - ErrInvalidConfig: VM is not in a startable state (e.g., already running, terminating)
	//   - ErrRateLimit: Too many requests, retry with exponential backoff
	//
	// Example:
	//   err := compute.StartVM(ctx, "i-1234567890abcdef0")
	//   if err != nil {
	//       log.Fatalf("Failed to start VM: %v", err)
	//   }
	//
	//   fmt.Println("VM start initiated. Waiting for running state...")
	//
	//   // Poll until running
	//   for {
	//       vm, err := compute.GetVM(ctx, "i-1234567890abcdef0")
	//       if err != nil {
	//           log.Fatalf("Failed to check VM status: %v", err)
	//       }
	//
	//       fmt.Printf("Current state: %s\n", vm.State)
	//       if vm.State == "running" {
	//           fmt.Printf("VM is now running! Public IP: %s\n", vm.PublicIP)
	//           break
	//       }
	//
	//       time.Sleep(10 * time.Second)
	//   }
	StartVM(ctx context.Context, id string) error

	// StopVM gracefully stops a running virtual machine.
	// The VM must be in "running" state. Stopping preserves the instance and its data.
	// The VM can be started again later with StartVM. You're not charged for compute time while stopped,
	// but storage charges may still apply.
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials or expired tokens
	//   - ErrAuthorization: Insufficient permissions or VM not owned by your account
	//   - ErrResourceNotFound: VM with the specified ID doesn't exist
	//   - ErrInvalidConfig: VM is not in a stoppable state (e.g., already stopped, terminating)
	//   - ErrRateLimit: Too many requests, retry with exponential backoff
	//
	// Example:
	//   err := compute.StopVM(ctx, "i-1234567890abcdef0")
	//   if err != nil {
	//       log.Fatalf("Failed to stop VM: %v", err)
	//   }
	//
	//   fmt.Println("VM stop initiated. Waiting for stopped state...")
	//
	//   // Poll until stopped
	//   for {
	//       vm, err := compute.GetVM(ctx, "i-1234567890abcdef0")
	//       if err != nil {
	//           log.Fatalf("Failed to check VM status: %v", err)
	//       }
	//
	//       fmt.Printf("Current state: %s\n", vm.State)
	//       if vm.State == "stopped" {
	//           fmt.Println("VM is now stopped. You can start it again later with StartVM.")
	//           break
	//       }
	//
	//       time.Sleep(10 * time.Second)
	//   }
	StopVM(ctx context.Context, id string) error

	// DeleteVM permanently deletes a virtual machine and all its associated data.
	// This operation cannot be undone. The VM and its ephemeral storage will be lost forever.
	// EBS volumes may be preserved depending on their DeleteOnTermination setting.
	//
	// The VM can be in any state, but "running" VMs will be stopped first before deletion.
	// Consider stopping the VM first if you want to ensure a graceful shutdown.
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials or expired tokens
	//   - ErrAuthorization: Insufficient permissions or VM not owned by your account
	//   - ErrResourceNotFound: VM with the specified ID doesn't exist
	//   - ErrResourceConflict: VM has termination protection enabled
	//   - ErrRateLimit: Too many requests, retry with exponential backoff
	//
	// Example:
	//   // Confirm deletion (this is permanent!)
	//   fmt.Print("Are you sure you want to delete this VM? (yes/no): ")
	//   var response string
	//   fmt.Scanln(&response)
	//   if response != "yes" {
	//       fmt.Println("Deletion cancelled")
	//       return
	//   }
	//
	//   err := compute.DeleteVM(ctx, "i-1234567890abcdef0")
	//   if err != nil {
	//       log.Fatalf("Failed to delete VM: %v", err)
	//   }
	//
	//   fmt.Println("VM deletion initiated. The VM will be terminated shortly.")
	//
	//   // Optional: Wait for termination
	//   for {
	//       vm, err := compute.GetVM(ctx, "i-1234567890abcdef0")
	//       if err != nil {
	//           if errors.Is(err, ErrResourceNotFound) {
	//               fmt.Println("VM has been successfully deleted")
	//               break
	//           }
	//           log.Fatalf("Failed to check VM status: %v", err)
	//       }
	//
	//       if vm.State == "terminated" {
	//           fmt.Println("VM has been terminated")
	//           break
	//       }
	//
	//       time.Sleep(10 * time.Second)
	//   }
	DeleteVM(ctx context.Context, id string) error

	// InstanceTypes returns the service for querying available instance types.
	// Use this to discover what hardware configurations are available in your region
	// and their specifications (CPU, memory, storage, network performance).
	//
	// Example:
	//   instanceTypes := compute.InstanceTypes()
	//   types, err := instanceTypes.List(ctx, nil)
	//   if err != nil {
	//       log.Fatalf("Failed to list instance types: %v", err)
	//   }
	//
	//   // Find suitable instance types for a web server
	//   for _, t := range types {
	//       if t.VCpus >= 2 && t.MemoryGB >= 4.0 && t.CurrentGeneration {
	//           fmt.Printf("Suitable: %s (%d vCPUs, %.1f GB RAM)\n",
	//               t.InstanceType, t.VCpus, t.MemoryGB)
	//       }
	//   }
	InstanceTypes() InstanceTypesService

	// PlacementGroups returns the service for managing placement groups.
	// Placement groups control how instances are positioned on underlying hardware
	// to optimize for network performance, fault tolerance, or both.
	//
	// Example:
	//   placementGroups := compute.PlacementGroups()
	//
	//   // Create a cluster placement group for high-performance computing
	//   config := &PlacementGroupConfig{
	//       GroupName: "hpc-cluster",
	//       Strategy:  "cluster",
	//   }
	//   group, err := placementGroups.Create(ctx, config)
	//   if err != nil {
	//       log.Fatalf("Failed to create placement group: %v", err)
	//   }
	//
	//   // Now launch instances into the placement group
	//   vmConfig := &VMConfig{
	//       Name:           "hpc-node-1",
	//       ImageID:        "ami-12345678",
	//       InstanceType:   "c5n.xlarge",
	//       PlacementGroup: group.GroupName,  // Add this field to VMConfig
	//   }
	PlacementGroups() PlacementGroupsService

	// SpotInstances returns the service for managing spot instances.
	// Spot instances offer significant cost savings (up to 90%) but can be interrupted
	// when capacity is needed for on-demand instances. Best for fault-tolerant workloads.
	//
	// Example:
	//   spotInstances := compute.SpotInstances()
	//
	//   // Request a spot instance for batch processing
	//   config := &SpotInstanceConfig{
	//       InstanceType: "m5.large",
	//       ImageID:      "ami-12345678",
	//       SpotPrice:    aws.String("0.10"),  // Max $0.10/hour (current on-demand: ~$0.096)
	//       LaunchSpecification: &SpotLaunchSpec{
	//           ImageID:        "ami-12345678",
	//           InstanceType:   "m5.large",
	//           KeyName:        "batch-processing-key",
	//           SecurityGroups: []string{"sg-batch"},
	//           UserData:       "#!/bin/bash\n# Setup batch processing environment",
	//       },
	//   }
	//
	//   request, err := spotInstances.Request(ctx, config)
	//   if err != nil {
	//       log.Fatalf("Failed to request spot instance: %v", err)
	//   }
	//
	//   fmt.Printf("Spot request created: %s\n", request.SpotInstanceRequestId)
	SpotInstances() SpotInstancesService
}
