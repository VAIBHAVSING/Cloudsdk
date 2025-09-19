package services

import (
	"context"
	"io"
)

// BucketConfig represents the configuration for creating a storage bucket.
// Buckets are containers for objects (files) and must have globally unique names.
// Supports JSON/YAML serialization for external configuration files.
//
// Bucket Naming Requirements (Strictly Enforced):
//   - Must be globally unique across ALL accounts and regions
//   - Length: 3-63 characters
//   - Lowercase letters, numbers, and hyphens only
//   - Must start and end with a letter or number
//   - Cannot contain consecutive hyphens (--)
//   - Cannot look like IP addresses (e.g., 192.168.1.1)
//   - Cannot start with "xn--" or end with "-s3alias"
//   - Cannot contain uppercase letters or underscores
//
// Regional Considerations:
//   - Choose regions close to your users for better performance
//   - Consider data residency and compliance requirements (GDPR, HIPAA)
//   - Some regions have different pricing tiers
//   - Cross-region replication may incur additional costs
//   - Availability varies by provider (not all services in all regions)
//
// Security and Compliance:
//   - Default to private access unless public access is specifically required
//   - Enable versioning for important data to protect against accidental deletion
//   - Consider encryption at rest for sensitive data
//   - Use bucket policies and IAM for fine-grained access control
//   - Enable access logging for audit trails
//   - Consider MFA delete for critical buckets
//
// Cost Optimization:
//   - Choose appropriate storage class based on access patterns
//   - Enable lifecycle policies to automatically transition or delete old objects
//   - Monitor storage usage and costs regularly
//   - Consider intelligent tiering for variable access patterns
//
// Example:
//
//	config := &BucketConfig{
//	    Name:       "mycompany-app-assets-prod-2024",  // Globally unique name
//	    Region:     "us-east-1",                       // Choose region close to users
//	    Versioning: true,                              // Keep multiple versions of objects
//	    ACL:        "private",                         // Default: private access only
//	    StorageClass: "STANDARD",                      // Standard storage for frequent access
//	    Encryption: &BucketEncryption{
//	        Enabled: true,
//	        KMSKeyID: "alias/my-bucket-key",
//	    },
//	    LifecycleRules: []LifecycleRule{
//	        {
//	            ID:     "archive-old-objects",
//	            Status: "Enabled",
//	            Transitions: []Transition{
//	                {Days: 30, StorageClass: "STANDARD_IA"},
//	                {Days: 90, StorageClass: "GLACIER"},
//	            },
//	        },
//	    },
//	    Tags: map[string]string{
//	        "Environment": "production",
//	        "Team":        "backend",
//	        "Project":     "web-app",
//	    },
//	}
type BucketConfig struct {
	// Name is the globally unique identifier for the bucket.
	// Once created, the name cannot be changed. Choose carefully.
	// Must follow strict naming conventions (see struct documentation above)
	//
	// Naming Best Practices:
	//   - Include organization name to avoid conflicts: "mycompany-app-data"
	//   - Use environment suffixes: "-prod", "-staging", "-dev"
	//   - Include purpose: "-assets", "-backups", "-logs"
	//   - Add date/version for temporary buckets: "-2024-01"
	//   - Avoid periods (can cause SSL certificate issues)
	//
	// Examples:
	//   - "mycompany-web-assets-prod"
	//   - "acme-corp-backups-us-east"
	//   - "startup-logs-staging-2024"
	Name string `json:"name" yaml:"name" validate:"required,min=3,max=63,lowercase,bucket_name"`

	// Region specifies where the bucket should be created.
	// Choose based on user proximity, compliance requirements, and cost
	// Leave empty to use the provider's default region
	//
	// Regional Considerations:
	//   - Latency: Choose regions close to your users
	//   - Compliance: Some data must stay in specific regions (GDPR, etc.)
	//   - Pricing: Costs vary significantly between regions
	//   - Features: Not all storage classes available in all regions
	//   - Disaster Recovery: Consider cross-region replication
	//
	// Common Regions:
	//   - AWS: us-east-1, us-west-2, eu-west-1, ap-southeast-1
	//   - GCP: us-central1, europe-west1, asia-east1
	//   - Azure: eastus, westeurope, southeastasia
	Region string `json:"region,omitempty" yaml:"region,omitempty"`

	// Versioning enables object versioning for the bucket.
	// When enabled, multiple versions of the same object can be stored
	// Provides protection against accidental deletion or modification
	//
	// Benefits:
	//   - Protect against accidental deletion or overwrite
	//   - Maintain history of object changes
	//   - Enable point-in-time recovery
	//   - Support for compliance requirements
	//
	// Considerations:
	//   - Increases storage costs (multiple versions stored)
	//   - Requires lifecycle policies to manage old versions
	//   - Delete operations create delete markers instead of removing objects
	//   - May complicate object management
	//
	// Best Practices:
	//   - Enable for critical data buckets
	//   - Use lifecycle policies to automatically delete old versions
	//   - Monitor storage costs when versioning is enabled
	//   - Consider MFA delete for additional protection
	//
	// Default: false (versioning disabled)
	Versioning *bool `json:"versioning,omitempty" yaml:"versioning,omitempty"`

	// ACL (Access Control List) sets the default access permissions.
	// Controls who can access the bucket and its objects
	// Use with caution - prefer IAM policies for fine-grained control
	//
	// Standard ACL Values:
	//   - "private": Only bucket owner has access (RECOMMENDED default)
	//   - "public-read": Public read access to all objects (USE WITH CAUTION)
	//   - "public-read-write": Public read/write access (NOT RECOMMENDED)
	//   - "authenticated-read": Read access for authenticated users only
	//   - "bucket-owner-read": Object owner and bucket owner have read access
	//   - "bucket-owner-full-control": Object owner and bucket owner have full control
	//
	// Security Warnings:
	//   - "public-read": Makes ALL objects publicly accessible on the internet
	//   - "public-read-write": Allows anyone to upload/modify objects (security risk)
	//   - Always audit public buckets for sensitive data
	//   - Consider using CloudFront or CDN instead of public buckets
	//
	// Best Practices:
	//   - Default to "private" unless public access is specifically required
	//   - Use IAM policies instead of ACLs for complex permissions
	//   - Regularly audit bucket permissions
	//   - Enable bucket notifications for permission changes
	//   - Use bucket policies to deny public access by default
	//
	// Default: "private"
	ACL string `json:"acl,omitempty" yaml:"acl,omitempty" validate:"oneof=private public-read public-read-write authenticated-read bucket-owner-read bucket-owner-full-control"`

	// StorageClass defines the storage tier for objects in the bucket.
	// Different classes offer different performance, availability, and cost characteristics
	// Can be overridden per object during upload
	//
	// AWS Storage Classes:
	//   - "STANDARD": Frequent access, high durability, low latency (default)
	//   - "STANDARD_IA": Infrequent access, lower cost, retrieval fees
	//   - "ONEZONE_IA": Infrequent access, single AZ, lowest cost
	//   - "GLACIER": Archive storage, minutes to hours retrieval
	//   - "DEEP_ARCHIVE": Long-term archive, 12+ hours retrieval
	//   - "INTELLIGENT_TIERING": Automatic tiering based on access patterns
	//
	// GCP Storage Classes:
	//   - "STANDARD": Frequent access, best performance
	//   - "NEARLINE": Monthly access, lower cost
	//   - "COLDLINE": Quarterly access, very low cost
	//   - "ARCHIVE": Annual access, lowest cost
	//
	// Azure Storage Tiers:
	//   - "Hot": Frequent access, higher storage cost, lower access cost
	//   - "Cool": Infrequent access, lower storage cost, higher access cost
	//   - "Archive": Rare access, lowest storage cost, highest access cost
	//
	// Selection Guidelines:
	//   - STANDARD: Active data, frequent access (websites, apps)
	//   - IA/NEARLINE/COOL: Backups, logs, infrequent access
	//   - GLACIER/COLDLINE/ARCHIVE: Long-term archival, compliance
	//   - INTELLIGENT_TIERING: Variable access patterns
	StorageClass string `json:"storage_class,omitempty" yaml:"storage_class,omitempty"`

	// Encryption configures server-side encryption for the bucket.
	// Encrypts objects at rest using provider-managed or customer-managed keys
	// Highly recommended for sensitive data and compliance requirements
	Encryption *BucketEncryption `json:"encryption,omitempty" yaml:"encryption,omitempty"`

	// LifecycleRules define automatic object management policies.
	// Automatically transition objects between storage classes or delete them
	// Essential for cost optimization and compliance
	LifecycleRules []LifecycleRule `json:"lifecycle_rules,omitempty" yaml:"lifecycle_rules,omitempty"`

	// Tags are key-value pairs for organizing and managing buckets.
	// Used for billing, automation, access control, and resource organization
	// Maximum 50 tags per bucket, each key/value max 255 characters
	//
	// Common Tag Patterns:
	//   - Environment: "production", "staging", "development"
	//   - Team: "backend", "frontend", "devops", "data"
	//   - Project: "web-app", "mobile-api", "data-pipeline"
	//   - Owner: "john.doe@company.com"
	//   - CostCenter: "engineering", "marketing", "sales"
	//   - DataClassification: "public", "internal", "confidential", "restricted"
	//   - Backup: "daily", "weekly", "none"
	//   - Compliance: "gdpr", "hipaa", "sox", "pci"
	//
	// Best Practices:
	//   - Use consistent naming conventions across organization
	//   - Include mandatory tags (Environment, Owner, Project)
	//   - Use tags for cost allocation and chargeback
	//   - Automate tag compliance with policies
	//   - Include data classification for security
	Tags map[string]string `json:"tags,omitempty" yaml:"tags,omitempty" validate:"max=50,dive,keys,max=255,endkeys,max=255"`

	// PublicAccessBlock prevents accidental public access to the bucket.
	// Provides an additional layer of security beyond ACLs and bucket policies
	// Recommended for all buckets unless public access is specifically required
	PublicAccessBlock *PublicAccessBlockConfig `json:"public_access_block,omitempty" yaml:"public_access_block,omitempty"`

	// NotificationConfig enables event notifications for bucket operations.
	// Triggers events when objects are created, deleted, or modified
	// Useful for automation, monitoring, and data processing pipelines
	NotificationConfig *BucketNotificationConfig `json:"notification_config,omitempty" yaml:"notification_config,omitempty"`

	// CorsRules define Cross-Origin Resource Sharing (CORS) policies.
	// Required for web applications that access bucket objects from browsers
	// Controls which domains can access bucket objects via JavaScript
	CorsRules []CorsRule `json:"cors_rules,omitempty" yaml:"cors_rules,omitempty"`

	// WebsiteConfig enables static website hosting for the bucket.
	// Allows serving static web content directly from the bucket
	// Useful for hosting single-page applications, documentation sites
	WebsiteConfig *WebsiteConfiguration `json:"website_config,omitempty" yaml:"website_config,omitempty"`

	// ReplicationConfig enables cross-region replication for the bucket.
	// Automatically replicates objects to another bucket in a different region
	// Provides disaster recovery and compliance benefits
	ReplicationConfig *ReplicationConfiguration `json:"replication_config,omitempty" yaml:"replication_config,omitempty"`
}

// BucketEncryption configures server-side encryption for bucket objects.
// Encrypts data at rest using provider-managed or customer-managed keys.
type BucketEncryption struct {
	// Enabled determines whether encryption is active for the bucket.
	// When true, all objects uploaded to the bucket will be encrypted
	// Default: false (no encryption)
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Algorithm specifies the encryption algorithm to use.
	// Common values: "AES256", "aws:kms", "google-kms", "azure-kms"
	// Default: "AES256" (provider-managed keys)
	Algorithm string `json:"algorithm,omitempty" yaml:"algorithm,omitempty"`

	// KMSKeyID specifies the customer-managed key for encryption.
	// Format varies by provider:
	//   - AWS: Key ID, ARN, or alias (e.g., "alias/my-key")
	//   - GCP: projects/PROJECT/locations/LOCATION/keyRings/RING/cryptoKeys/KEY
	//   - Azure: Key vault key identifier
	// Leave empty to use provider-managed keys
	KMSKeyID string `json:"kms_key_id,omitempty" yaml:"kms_key_id,omitempty"`
}

// LifecycleRule defines automatic object management policies.
// Rules can transition objects between storage classes or delete them based on age.
type LifecycleRule struct {
	// ID is a unique identifier for the lifecycle rule.
	// Must be unique within the bucket
	// Used for management and troubleshooting
	ID string `json:"id" yaml:"id" validate:"required,max=255"`

	// Status determines whether the rule is active.
	// Values: "Enabled", "Disabled"
	// Disabled rules are ignored but preserved
	Status string `json:"status" yaml:"status" validate:"required,oneof=Enabled Disabled"`

	// Filter specifies which objects the rule applies to.
	// Can filter by prefix, tags, or object size
	Filter *LifecycleFilter `json:"filter,omitempty" yaml:"filter,omitempty"`

	// Transitions define storage class changes based on object age.
	// Objects automatically move to cheaper storage classes over time
	// Must be in chronological order (earliest first)
	Transitions []Transition `json:"transitions,omitempty" yaml:"transitions,omitempty"`

	// Expiration defines when objects should be deleted.
	// Automatically deletes objects after specified time period
	// Use with caution - deleted objects cannot be recovered
	Expiration *Expiration `json:"expiration,omitempty" yaml:"expiration,omitempty"`
}

// LifecycleFilter specifies which objects a lifecycle rule applies to.
type LifecycleFilter struct {
	// Prefix filters objects by key prefix (folder-like structure).
	// Example: "logs/" applies to all objects starting with "logs/"
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty"`

	// Tags filter objects by their tags.
	// All specified tags must match for the rule to apply
	Tags map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`

	// ObjectSizeGreaterThan filters objects larger than specified size in bytes.
	// Useful for applying different rules to large vs small objects
	ObjectSizeGreaterThan *int64 `json:"object_size_greater_than,omitempty" yaml:"object_size_greater_than,omitempty"`

	// ObjectSizeLessThan filters objects smaller than specified size in bytes.
	// Useful for applying different rules to large vs small objects
	ObjectSizeLessThan *int64 `json:"object_size_less_than,omitempty" yaml:"object_size_less_than,omitempty"`
}

// Transition defines a storage class change based on object age.
type Transition struct {
	// Days specifies the number of days after object creation.
	// Must be positive integer, minimum varies by storage class
	Days int32 `json:"days" yaml:"days" validate:"required,min=1"`

	// StorageClass is the target storage class for the transition.
	// Must be a valid storage class for the provider
	// Examples: "STANDARD_IA", "GLACIER", "DEEP_ARCHIVE"
	StorageClass string `json:"storage_class" yaml:"storage_class" validate:"required"`
}

// Expiration defines when objects should be automatically deleted.
type Expiration struct {
	// Days specifies the number of days after object creation.
	// Objects will be permanently deleted after this period
	// Use with extreme caution - deletion cannot be undone
	Days int32 `json:"days" yaml:"days" validate:"required,min=1"`

	// ExpiredObjectDeleteMarker removes delete markers for expired objects.
	// Only applies when versioning is enabled
	// Helps clean up delete markers left by expired objects
	ExpiredObjectDeleteMarker *bool `json:"expired_object_delete_marker,omitempty" yaml:"expired_object_delete_marker,omitempty"`
}

// PublicAccessBlockConfig prevents accidental public access to buckets.
// Provides additional security layer beyond ACLs and bucket policies.
type PublicAccessBlockConfig struct {
	// BlockPublicAcls blocks new public ACLs and uploading objects with public ACLs.
	// Recommended: true (prevents accidental public access)
	BlockPublicAcls *bool `json:"block_public_acls,omitempty" yaml:"block_public_acls,omitempty"`

	// IgnorePublicAcls ignores existing public ACLs on bucket and objects.
	// Recommended: true (treats public ACLs as private)
	IgnorePublicAcls *bool `json:"ignore_public_acls,omitempty" yaml:"ignore_public_acls,omitempty"`

	// BlockPublicPolicy blocks new public bucket policies.
	// Recommended: true (prevents public bucket policies)
	BlockPublicPolicy *bool `json:"block_public_policy,omitempty" yaml:"block_public_policy,omitempty"`

	// RestrictPublicBuckets restricts access to buckets with public policies.
	// Recommended: true (limits access to authorized users only)
	RestrictPublicBuckets *bool `json:"restrict_public_buckets,omitempty" yaml:"restrict_public_buckets,omitempty"`
}

// BucketNotificationConfig enables event notifications for bucket operations.
type BucketNotificationConfig struct {
	// TopicConfigurations define SNS topic notifications.
	// Sends messages to SNS topics when specified events occur
	TopicConfigurations []TopicConfiguration `json:"topic_configurations,omitempty" yaml:"topic_configurations,omitempty"`

	// QueueConfigurations define SQS queue notifications.
	// Sends messages to SQS queues when specified events occur
	QueueConfigurations []QueueConfiguration `json:"queue_configurations,omitempty" yaml:"queue_configurations,omitempty"`

	// LambdaConfigurations define Lambda function notifications.
	// Invokes Lambda functions when specified events occur
	LambdaConfigurations []LambdaConfiguration `json:"lambda_configurations,omitempty" yaml:"lambda_configurations,omitempty"`
}

// TopicConfiguration defines SNS topic notification settings.
type TopicConfiguration struct {
	// TopicArn is the ARN of the SNS topic to notify.
	TopicArn string `json:"topic_arn" yaml:"topic_arn" validate:"required"`

	// Events specify which bucket events trigger notifications.
	// Examples: "s3:ObjectCreated:*", "s3:ObjectRemoved:*"
	Events []string `json:"events" yaml:"events" validate:"required,min=1"`

	// Filter specifies which objects trigger notifications.
	Filter *NotificationFilter `json:"filter,omitempty" yaml:"filter,omitempty"`
}

// QueueConfiguration defines SQS queue notification settings.
type QueueConfiguration struct {
	// QueueArn is the ARN of the SQS queue to notify.
	QueueArn string `json:"queue_arn" yaml:"queue_arn" validate:"required"`

	// Events specify which bucket events trigger notifications.
	Events []string `json:"events" yaml:"events" validate:"required,min=1"`

	// Filter specifies which objects trigger notifications.
	Filter *NotificationFilter `json:"filter,omitempty" yaml:"filter,omitempty"`
}

// LambdaConfiguration defines Lambda function notification settings.
type LambdaConfiguration struct {
	// LambdaFunctionArn is the ARN of the Lambda function to invoke.
	LambdaFunctionArn string `json:"lambda_function_arn" yaml:"lambda_function_arn" validate:"required"`

	// Events specify which bucket events trigger notifications.
	Events []string `json:"events" yaml:"events" validate:"required,min=1"`

	// Filter specifies which objects trigger notifications.
	Filter *NotificationFilter `json:"filter,omitempty" yaml:"filter,omitempty"`
}

// NotificationFilter specifies which objects trigger notifications.
type NotificationFilter struct {
	// Key filters notifications by object key (name).
	Key *KeyFilter `json:"key,omitempty" yaml:"key,omitempty"`
}

// KeyFilter filters notifications by object key patterns.
type KeyFilter struct {
	// FilterRules define prefix and suffix matching rules.
	FilterRules []FilterRule `json:"filter_rules,omitempty" yaml:"filter_rules,omitempty"`
}

// FilterRule defines a single key filtering rule.
type FilterRule struct {
	// Name specifies the filter type: "prefix" or "suffix".
	Name string `json:"name" yaml:"name" validate:"required,oneof=prefix suffix"`

	// Value is the string to match against object keys.
	Value string `json:"value" yaml:"value" validate:"required"`
}

// CorsRule defines Cross-Origin Resource Sharing (CORS) policies.
type CorsRule struct {
	// ID is a unique identifier for the CORS rule.
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// AllowedHeaders specify which headers can be used in requests.
	// Use ["*"] to allow all headers
	AllowedHeaders []string `json:"allowed_headers,omitempty" yaml:"allowed_headers,omitempty"`

	// AllowedMethods specify which HTTP methods are allowed.
	// Common values: "GET", "PUT", "POST", "DELETE", "HEAD"
	AllowedMethods []string `json:"allowed_methods" yaml:"allowed_methods" validate:"required,min=1"`

	// AllowedOrigins specify which domains can make requests.
	// Use ["*"] to allow all origins (not recommended for production)
	// Examples: ["https://example.com", "https://*.example.com"]
	AllowedOrigins []string `json:"allowed_origins" yaml:"allowed_origins" validate:"required,min=1"`

	// ExposeHeaders specify which headers browsers can access.
	// Examples: ["ETag", "x-amz-request-id"]
	ExposeHeaders []string `json:"expose_headers,omitempty" yaml:"expose_headers,omitempty"`

	// MaxAgeSeconds specifies how long browsers can cache preflight responses.
	// Default: 3600 seconds (1 hour)
	MaxAgeSeconds *int32 `json:"max_age_seconds,omitempty" yaml:"max_age_seconds,omitempty"`
}

// WebsiteConfiguration enables static website hosting.
type WebsiteConfiguration struct {
	// IndexDocument specifies the default page for the website.
	// Common values: "index.html", "index.htm"
	IndexDocument string `json:"index_document" yaml:"index_document" validate:"required"`

	// ErrorDocument specifies the error page for the website.
	// Common values: "error.html", "404.html"
	ErrorDocument string `json:"error_document,omitempty" yaml:"error_document,omitempty"`

	// RedirectAllRequestsTo redirects all requests to another host.
	// Useful for domain redirects or maintenance pages
	RedirectAllRequestsTo *RedirectAllRequestsTo `json:"redirect_all_requests_to,omitempty" yaml:"redirect_all_requests_to,omitempty"`

	// RoutingRules define advanced routing for the website.
	// Allows conditional redirects based on request properties
	RoutingRules []RoutingRule `json:"routing_rules,omitempty" yaml:"routing_rules,omitempty"`
}

// RedirectAllRequestsTo defines a global redirect for all requests.
type RedirectAllRequestsTo struct {
	// HostName is the target host for redirects.
	// Example: "example.com"
	HostName string `json:"host_name" yaml:"host_name" validate:"required"`

	// Protocol specifies the redirect protocol.
	// Values: "http", "https"
	Protocol string `json:"protocol,omitempty" yaml:"protocol,omitempty" validate:"omitempty,oneof=http https"`
}

// RoutingRule defines conditional redirects for website hosting.
type RoutingRule struct {
	// Condition specifies when the rule applies.
	Condition *RoutingRuleCondition `json:"condition,omitempty" yaml:"condition,omitempty"`

	// Redirect specifies the redirect behavior.
	Redirect RoutingRuleRedirect `json:"redirect" yaml:"redirect" validate:"required"`
}

// RoutingRuleCondition defines when a routing rule applies.
type RoutingRuleCondition struct {
	// KeyPrefixEquals matches requests with specific key prefixes.
	// Example: "documents/" matches all requests starting with "documents/"
	KeyPrefixEquals string `json:"key_prefix_equals,omitempty" yaml:"key_prefix_equals,omitempty"`

	// HttpErrorCodeReturnedEquals matches specific HTTP error codes.
	// Example: "404" matches all 404 Not Found errors
	HttpErrorCodeReturnedEquals string `json:"http_error_code_returned_equals,omitempty" yaml:"http_error_code_returned_equals,omitempty"`
}

// RoutingRuleRedirect defines redirect behavior for routing rules.
type RoutingRuleRedirect struct {
	// HostName is the target host for the redirect.
	HostName string `json:"host_name,omitempty" yaml:"host_name,omitempty"`

	// HttpRedirectCode specifies the HTTP redirect status code.
	// Common values: "301" (permanent), "302" (temporary)
	HttpRedirectCode string `json:"http_redirect_code,omitempty" yaml:"http_redirect_code,omitempty"`

	// Protocol specifies the redirect protocol.
	Protocol string `json:"protocol,omitempty" yaml:"protocol,omitempty" validate:"omitempty,oneof=http https"`

	// ReplaceKeyPrefixWith replaces the key prefix in redirects.
	ReplaceKeyPrefixWith string `json:"replace_key_prefix_with,omitempty" yaml:"replace_key_prefix_with,omitempty"`

	// ReplaceKeyWith replaces the entire key in redirects.
	ReplaceKeyWith string `json:"replace_key_with,omitempty" yaml:"replace_key_with,omitempty"`
}

// ReplicationConfiguration enables cross-region replication.
type ReplicationConfiguration struct {
	// Role is the IAM role ARN for replication permissions.
	// Must have permissions to read from source and write to destination
	Role string `json:"role" yaml:"role" validate:"required"`

	// Rules define replication behavior for different object sets.
	Rules []ReplicationRule `json:"rules" yaml:"rules" validate:"required,min=1"`
}

// ReplicationRule defines replication behavior for a set of objects.
type ReplicationRule struct {
	// ID is a unique identifier for the replication rule.
	ID string `json:"id" yaml:"id" validate:"required,max=255"`

	// Status determines whether the rule is active.
	// Values: "Enabled", "Disabled"
	Status string `json:"status" yaml:"status" validate:"required,oneof=Enabled Disabled"`

	// Priority determines rule precedence when multiple rules apply.
	// Higher numbers have higher priority
	Priority int32 `json:"priority,omitempty" yaml:"priority,omitempty"`

	// Filter specifies which objects to replicate.
	Filter *ReplicationRuleFilter `json:"filter,omitempty" yaml:"filter,omitempty"`

	// Destination specifies where to replicate objects.
	Destination ReplicationDestination `json:"destination" yaml:"destination" validate:"required"`

	// DeleteMarkerReplication controls replication of delete markers.
	DeleteMarkerReplication *DeleteMarkerReplication `json:"delete_marker_replication,omitempty" yaml:"delete_marker_replication,omitempty"`
}

// ReplicationRuleFilter specifies which objects to replicate.
type ReplicationRuleFilter struct {
	// Prefix filters objects by key prefix.
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty"`

	// Tags filter objects by their tags.
	Tags map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// ReplicationDestination specifies where to replicate objects.
type ReplicationDestination struct {
	// Bucket is the ARN of the destination bucket.
	Bucket string `json:"bucket" yaml:"bucket" validate:"required"`

	// StorageClass specifies the storage class for replicated objects.
	// Leave empty to use the same class as source objects
	StorageClass string `json:"storage_class,omitempty" yaml:"storage_class,omitempty"`

	// EncryptionConfiguration specifies encryption for replicated objects.
	EncryptionConfiguration *ReplicationEncryptionConfiguration `json:"encryption_configuration,omitempty" yaml:"encryption_configuration,omitempty"`
}

// ReplicationEncryptionConfiguration specifies encryption for replicated objects.
type ReplicationEncryptionConfiguration struct {
	// ReplicaKmsKeyID is the KMS key ID for encrypting replicated objects.
	ReplicaKmsKeyID string `json:"replica_kms_key_id" yaml:"replica_kms_key_id" validate:"required"`
}

// DeleteMarkerReplication controls replication of delete markers.
type DeleteMarkerReplication struct {
	// Status determines whether delete markers are replicated.
	// Values: "Enabled", "Disabled"
	Status string `json:"status" yaml:"status" validate:"required,oneof=Enabled Disabled"`
}

// Object represents a file stored in a bucket with its metadata.
// Objects are the fundamental entities stored in cloud storage services.
type Object struct {
	// Key is the unique identifier for the object within the bucket.
	// This is essentially the file path/name within the bucket.
	// Can include forward slashes to simulate folder structures.
	// Examples: "images/photo.jpg", "documents/2024/report.pdf", "data.json"
	Key string

	// Size is the object size in bytes.
	// This is the actual size of the stored data, not including metadata.
	// Use this to calculate storage costs or validate file sizes.
	Size int64

	// LastModified indicates when the object was last updated.
	// Format: RFC3339 timestamp (e.g., "2023-01-15T10:30:00Z")
	// Updated whenever the object content is replaced.
	LastModified string

	// ETag is a hash of the object content used for integrity checking.
	// Format varies by provider but typically looks like an MD5 hash.
	// Changes whenever the object content changes.
	// Useful for detecting if an object has been modified.
	ETag string
}

// Storage provides object storage operations across cloud providers.
// This interface abstracts the differences between AWS S3, Google Cloud Storage,
// Azure Blob Storage, and other object storage services.
//
// Security Considerations:
//   - Always use HTTPS for data transfer (handled automatically by the SDK)
//   - Consider encryption at rest for sensitive data
//   - Use IAM policies and bucket policies to control access
//   - Regularly audit bucket permissions and access logs
//   - Enable versioning for important data to protect against accidental deletion
//
// Cost Optimization:
//   - Choose the appropriate storage class for your access patterns
//   - Use lifecycle policies to automatically transition or delete old objects
//   - Monitor storage usage and costs regularly
//   - Consider compression for large text files
//
// Performance Best Practices:
//   - Use multipart uploads for large files (>100MB)
//   - Implement retry logic with exponential backoff
//   - Use CloudFront or CDN for frequently accessed content
//   - Choose regions close to your users
//
// Bucket Naming Best Practices:
//   - Include your organization name to avoid conflicts
//   - Use environment suffixes (e.g., "-prod", "-staging")
//   - Avoid using periods in bucket names (can cause SSL certificate issues)
//   - Use consistent naming conventions across your organization
type Storage interface {
	// CreateBucket creates a new storage bucket with the specified configuration.
	// Bucket names must be globally unique across all accounts and regions.
	// The bucket will be created in the specified region or the default region.
	//
	// Bucket Naming Validation:
	//   - Length: 3-63 characters
	//   - Characters: lowercase letters, numbers, hyphens only
	//   - Must start and end with letter or number
	//   - Cannot contain consecutive hyphens
	//   - Cannot look like IP addresses (e.g., 192.168.1.1)
	//
	// Common errors:
	//   - ErrResourceConflict: Bucket name already exists globally
	//   - ErrInvalidConfig: Invalid bucket name or configuration
	//   - ErrAuthorization: Insufficient permissions to create buckets
	//   - ErrRateLimit: Too many bucket creation requests
	//
	// Example:
	//   config := &BucketConfig{
	//       Name:       "mycompany-app-assets-prod",
	//       Region:     "us-east-1",
	//       Versioning: true,    // Enable versioning for data protection
	//       ACL:        "private", // Keep bucket private by default
	//   }
	//
	//   err := storage.CreateBucket(ctx, config)
	//   if err != nil {
	//       if errors.Is(err, ErrResourceConflict) {
	//           fmt.Println("Bucket name already exists. Try a different name.")
	//           return
	//       }
	//       log.Fatalf("Failed to create bucket: %v", err)
	//   }
	//
	//   fmt.Printf("Bucket '%s' created successfully\n", config.Name)
	CreateBucket(ctx context.Context, config *BucketConfig) error

	// ListBuckets returns the names of all buckets in your account.
	// Returns an empty slice if no buckets exist.
	// Note: This operation lists buckets across all regions.
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials or expired tokens
	//   - ErrAuthorization: Insufficient permissions to list buckets
	//   - ErrRateLimit: Too many list requests
	//
	// Example:
	//   buckets, err := storage.ListBuckets(ctx)
	//   if err != nil {
	//       log.Fatalf("Failed to list buckets: %v", err)
	//   }
	//
	//   if len(buckets) == 0 {
	//       fmt.Println("No buckets found. Create one with CreateBucket().")
	//       return
	//   }
	//
	//   fmt.Printf("Found %d buckets:\n", len(buckets))
	//   for i, bucket := range buckets {
	//       fmt.Printf("  %d. %s\n", i+1, bucket)
	//   }
	//
	//   // Find buckets by environment
	//   prodBuckets := make([]string, 0)
	//   for _, bucket := range buckets {
	//       if strings.Contains(bucket, "-prod") {
	//           prodBuckets = append(prodBuckets, bucket)
	//       }
	//   }
	//   fmt.Printf("Production buckets: %d\n", len(prodBuckets))
	ListBuckets(ctx context.Context) ([]string, error)

	// DeleteBucket permanently deletes a bucket and all its contents.
	// This operation cannot be undone. All objects in the bucket will be lost forever.
	// The bucket must be empty before deletion (use ListObjects and DeleteObject first).
	//
	// Important: Some providers require the bucket to be completely empty,
	// including all object versions and incomplete multipart uploads.
	//
	// Common errors:
	//   - ErrResourceNotFound: Bucket doesn't exist
	//   - ErrResourceConflict: Bucket is not empty
	//   - ErrAuthorization: Insufficient permissions to delete bucket
	//   - ErrInvalidConfig: Bucket has deletion protection enabled
	//
	// Example:
	//   bucketName := "old-test-bucket"
	//
	//   // First, list and delete all objects
	//   objects, err := storage.ListObjects(ctx, bucketName)
	//   if err != nil {
	//       log.Fatalf("Failed to list objects: %v", err)
	//   }
	//
	//   fmt.Printf("Deleting %d objects from bucket...\n", len(objects))
	//   for _, obj := range objects {
	//       err := storage.DeleteObject(ctx, bucketName, obj.Key)
	//       if err != nil {
	//           log.Printf("Failed to delete object %s: %v", obj.Key, err)
	//       }
	//   }
	//
	//   // Now delete the empty bucket
	//   err = storage.DeleteBucket(ctx, bucketName)
	//   if err != nil {
	//       log.Fatalf("Failed to delete bucket: %v", err)
	//   }
	//
	//   fmt.Printf("Bucket '%s' deleted successfully\n", bucketName)
	DeleteBucket(ctx context.Context, name string) error

	// PutObject uploads data to the specified bucket and key (file path).
	// If an object with the same key already exists, it will be overwritten.
	// The data is read from the provided io.Reader until EOF.
	//
	// Upload Considerations:
	//   - For files >100MB, consider using multipart uploads (provider-specific)
	//   - The entire content is read into memory, so be mindful of large files
	//   - Content-Type is automatically detected based on file extension
	//   - Data is encrypted in transit (HTTPS) and optionally at rest
	//
	// Key (File Path) Guidelines:
	//   - Use forward slashes for folder-like organization: "images/2024/photo.jpg"
	//   - Avoid leading slashes: use "folder/file.txt" not "/folder/file.txt"
	//   - Use consistent naming conventions for easier management
	//   - Consider including timestamps or UUIDs for uniqueness
	//
	// Common errors:
	//   - ErrResourceNotFound: Bucket doesn't exist
	//   - ErrAuthorization: Insufficient permissions to write to bucket
	//   - ErrInvalidConfig: Invalid key name or bucket configuration
	//   - ErrRateLimit: Too many upload requests
	//   - ErrNetworkTimeout: Upload timeout (retry with exponential backoff)
	//
	// Example:
	//   // Upload a file from disk
	//   file, err := os.Open("photo.jpg")
	//   if err != nil {
	//       log.Fatalf("Failed to open file: %v", err)
	//   }
	//   defer file.Close()
	//
	//   err = storage.PutObject(ctx, "my-photos", "vacation/2024/beach.jpg", file)
	//   if err != nil {
	//       log.Fatalf("Failed to upload file: %v", err)
	//   }
	//
	//   fmt.Println("File uploaded successfully!")
	//
	//   // Upload data from memory
	//   data := strings.NewReader(`{"message": "Hello, World!", "timestamp": "2024-01-15T10:30:00Z"}`)
	//   err = storage.PutObject(ctx, "my-data", "messages/hello.json", data)
	//   if err != nil {
	//       log.Fatalf("Failed to upload data: %v", err)
	//   }
	//
	//   // Upload with progress tracking (for large files)
	//   largeFile, _ := os.Open("large-video.mp4")
	//   defer largeFile.Close()
	//
	//   // Wrap reader to track progress
	//   fileInfo, _ := largeFile.Stat()
	//   progressReader := &ProgressReader{
	//       Reader: largeFile,
	//       Total:  fileInfo.Size(),
	//       OnProgress: func(bytesRead, total int64) {
	//           percent := float64(bytesRead) / float64(total) * 100
	//           fmt.Printf("\rUpload progress: %.1f%%", percent)
	//       },
	//   }
	//
	//   err = storage.PutObject(ctx, "my-videos", "uploads/video.mp4", progressReader)
	PutObject(ctx context.Context, bucket, key string, body io.Reader) error

	// GetObject downloads an object from the specified bucket and key.
	// Returns a ReadCloser that streams the object content.
	// The caller MUST close the returned ReadCloser to avoid resource leaks.
	//
	// Download Considerations:
	//   - The returned ReadCloser streams data, so it's memory-efficient for large files
	//   - Always close the ReadCloser when done to free network connections
	//   - For large files, consider implementing retry logic for network interruptions
	//   - The object is downloaded over HTTPS for security
	//
	// Common errors:
	//   - ErrResourceNotFound: Bucket or object doesn't exist
	//   - ErrAuthorization: Insufficient permissions to read from bucket
	//   - ErrNetworkTimeout: Download timeout (retry with exponential backoff)
	//   - ErrRateLimit: Too many download requests
	//
	// Example:
	//   // Download and save to file
	//   reader, err := storage.GetObject(ctx, "my-photos", "vacation/2024/beach.jpg")
	//   if err != nil {
	//       if errors.Is(err, ErrResourceNotFound) {
	//           fmt.Println("Photo not found")
	//           return
	//       }
	//       log.Fatalf("Failed to download photo: %v", err)
	//   }
	//   defer reader.Close() // Important: always close!
	//
	//   outFile, err := os.Create("downloaded-beach.jpg")
	//   if err != nil {
	//       log.Fatalf("Failed to create output file: %v", err)
	//   }
	//   defer outFile.Close()
	//
	//   bytesWritten, err := io.Copy(outFile, reader)
	//   if err != nil {
	//       log.Fatalf("Failed to save file: %v", err)
	//   }
	//
	//   fmt.Printf("Downloaded %d bytes to downloaded-beach.jpg\n", bytesWritten)
	//
	//   // Download JSON data and parse
	//   reader, err = storage.GetObject(ctx, "my-data", "config/app-settings.json")
	//   if err != nil {
	//       log.Fatalf("Failed to download config: %v", err)
	//   }
	//   defer reader.Close()
	//
	//   var config map[string]interface{}
	//   decoder := json.NewDecoder(reader)
	//   err = decoder.Decode(&config)
	//   if err != nil {
	//       log.Fatalf("Failed to parse JSON: %v", err)
	//   }
	//
	//   fmt.Printf("Config loaded: %+v\n", config)
	//
	//   // Download with progress tracking
	//   reader, err = storage.GetObject(ctx, "my-videos", "large-file.mp4")
	//   if err != nil {
	//       log.Fatalf("Failed to start download: %v", err)
	//   }
	//   defer reader.Close()
	//
	//   // Note: You'll need to implement ProgressWriter similar to ProgressReader
	//   outFile, _ = os.Create("large-file.mp4")
	//   defer outFile.Close()
	//
	//   _, err = io.Copy(outFile, reader)
	//   if err != nil {
	//       log.Fatalf("Download failed: %v", err)
	//   }
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)

	// DeleteObject removes an object from the specified bucket.
	// This operation cannot be undone unless versioning is enabled on the bucket.
	// If versioning is enabled, this creates a delete marker instead of permanently deleting.
	//
	// Versioning Behavior:
	//   - Without versioning: Object is permanently deleted
	//   - With versioning: A delete marker is added, but previous versions remain
	//   - To permanently delete versioned objects, you need provider-specific APIs
	//
	// Common errors:
	//   - ErrResourceNotFound: Bucket or object doesn't exist (may not be an error)
	//   - ErrAuthorization: Insufficient permissions to delete from bucket
	//   - ErrRateLimit: Too many delete requests
	//
	// Example:
	//   // Delete a single object
	//   err := storage.DeleteObject(ctx, "my-photos", "old-photo.jpg")
	//   if err != nil {
	//       if errors.Is(err, ErrResourceNotFound) {
	//           fmt.Println("Object already deleted or doesn't exist")
	//       } else {
	//           log.Fatalf("Failed to delete object: %v", err)
	//       }
	//   } else {
	//       fmt.Println("Object deleted successfully")
	//   }
	//
	//   // Delete multiple objects (batch operation)
	//   objectsToDelete := []string{
	//       "temp/file1.txt",
	//       "temp/file2.txt",
	//       "temp/file3.txt",
	//   }
	//
	//   fmt.Printf("Deleting %d objects...\n", len(objectsToDelete))
	//   for i, key := range objectsToDelete {
	//       err := storage.DeleteObject(ctx, "my-bucket", key)
	//       if err != nil {
	//           log.Printf("Failed to delete %s: %v", key, err)
	//       } else {
	//           fmt.Printf("Deleted %d/%d: %s\n", i+1, len(objectsToDelete), key)
	//       }
	//   }
	//
	//   // Delete all objects with a prefix (simulate folder deletion)
	//   objects, err := storage.ListObjects(ctx, "my-bucket")
	//   if err != nil {
	//       log.Fatalf("Failed to list objects: %v", err)
	//   }
	//
	//   prefix := "temp/"
	//   deletedCount := 0
	//   for _, obj := range objects {
	//       if strings.HasPrefix(obj.Key, prefix) {
	//           err := storage.DeleteObject(ctx, "my-bucket", obj.Key)
	//           if err != nil {
	//               log.Printf("Failed to delete %s: %v", obj.Key, err)
	//           } else {
	//               deletedCount++
	//           }
	//       }
	//   }
	//   fmt.Printf("Deleted %d objects with prefix '%s'\n", deletedCount, prefix)
	DeleteObject(ctx context.Context, bucket, key string) error

	// ListObjects returns all objects in the specified bucket with their metadata.
	// Returns an empty slice if the bucket is empty.
	// Note: This operation may be expensive for buckets with many objects.
	//
	// Performance Considerations:
	//   - For buckets with >1000 objects, consider using provider-specific pagination
	//   - Results are not sorted by default - sort client-side if needed
	//   - Large buckets may take time to list completely
	//   - Consider using prefix-based filtering for better performance
	//
	// Object Metadata:
	//   - Key: The object's path/name within the bucket
	//   - Size: Object size in bytes (useful for storage cost calculation)
	//   - LastModified: When the object was last updated
	//   - ETag: Content hash for integrity verification
	//
	// Common errors:
	//   - ErrResourceNotFound: Bucket doesn't exist
	//   - ErrAuthorization: Insufficient permissions to list bucket contents
	//   - ErrRateLimit: Too many list requests
	//
	// Example:
	//   objects, err := storage.ListObjects(ctx, "my-photos")
	//   if err != nil {
	//       log.Fatalf("Failed to list objects: %v", err)
	//   }
	//
	//   if len(objects) == 0 {
	//       fmt.Println("Bucket is empty")
	//       return
	//   }
	//
	//   fmt.Printf("Found %d objects:\n", len(objects))
	//
	//   var totalSize int64
	//   for _, obj := range objects {
	//       fmt.Printf("  %s (%d bytes, modified: %s)\n",
	//           obj.Key, obj.Size, obj.LastModified)
	//       totalSize += obj.Size
	//   }
	//
	//   fmt.Printf("Total storage used: %.2f MB\n", float64(totalSize)/(1024*1024))
	//
	//   // Find objects by extension
	//   imageObjects := make([]*Object, 0)
	//   for _, obj := range objects {
	//       ext := strings.ToLower(filepath.Ext(obj.Key))
	//       if ext == ".jpg" || ext == ".png" || ext == ".gif" {
	//           imageObjects = append(imageObjects, obj)
	//       }
	//   }
	//   fmt.Printf("Image files: %d\n", len(imageObjects))
	//
	//   // Find recently modified objects (last 24 hours)
	//   yesterday := time.Now().Add(-24 * time.Hour)
	//   recentObjects := make([]*Object, 0)
	//   for _, obj := range objects {
	//       modTime, err := time.Parse(time.RFC3339, obj.LastModified)
	//       if err == nil && modTime.After(yesterday) {
	//           recentObjects = append(recentObjects, obj)
	//       }
	//   }
	//   fmt.Printf("Recently modified objects: %d\n", len(recentObjects))
	//
	//   // Sort objects by size (largest first)
	//   sort.Slice(objects, func(i, j int) bool {
	//       return objects[i].Size > objects[j].Size
	//   })
	//
	//   fmt.Println("Largest objects:")
	//   for i, obj := range objects {
	//       if i >= 5 { break } // Show top 5
	//       fmt.Printf("  %s: %.2f MB\n", obj.Key, float64(obj.Size)/(1024*1024))
	//   }
	ListObjects(ctx context.Context, bucket string) ([]*Object, error)
}
