package services

import "context"

// DBConfig represents the configuration for creating a managed database instance.
// Managed databases provide automated backups, patching, monitoring, and high availability.
// Supports JSON/YAML serialization for external configuration files.
//
// Supported Database Engines and Versions:
//
// PostgreSQL (Recommended for new applications):
//   - Versions: 11.x, 12.x, 13.x, 14.x, 15.x, 16.x
//   - Features: Advanced SQL, JSON support, full-text search, extensions
//   - Use cases: Web applications, analytics, geospatial data
//   - Port: 5432
//
// MySQL (Popular for web applications):
//   - Versions: 5.7.x, 8.0.x, 8.1.x
//   - Features: High performance, replication, partitioning
//   - Use cases: Web applications, e-commerce, content management
//   - Port: 3306
//
// MariaDB (MySQL-compatible with additional features):
//   - Versions: 10.4.x, 10.5.x, 10.6.x, 10.11.x
//   - Features: Enhanced performance, additional storage engines
//   - Use cases: Drop-in MySQL replacement, high-performance applications
//   - Port: 3306
//
// Oracle Database (Enterprise applications):
//   - Editions: Standard Edition 2 (SE2), Enterprise Edition (EE)
//   - Versions: 19c, 21c
//   - Features: Advanced analytics, high availability, security
//   - Use cases: Enterprise applications, data warehousing
//   - Port: 1521
//
// SQL Server (Microsoft ecosystem):
//   - Editions: Express, Web, Standard, Enterprise
//   - Versions: 2017, 2019, 2022
//   - Features: Integration with Microsoft tools, business intelligence
//   - Use cases: .NET applications, business intelligence, enterprise apps
//   - Port: 1433
//
// Instance Sizing Recommendations:
//
// Development/Testing:
//   - AWS: db.t3.micro (1 vCPU, 1GB RAM) - Free tier eligible
//   - GCP: db-f1-micro (0.6GB RAM) - Shared CPU
//   - Azure: Basic (1 vCore, 2GB RAM)
//
// Small Production (< 100 concurrent users):
//   - AWS: db.t3.small (2 vCPU, 2GB RAM) or db.t3.medium (2 vCPU, 4GB RAM)
//   - GCP: db-g1-small (1.7GB RAM) or db-n1-standard-1 (3.75GB RAM)
//   - Azure: GeneralPurpose (2 vCore, 10.2GB RAM)
//
// Medium Production (100-1000 concurrent users):
//   - AWS: db.r5.large (2 vCPU, 16GB RAM) or db.r5.xlarge (4 vCPU, 32GB RAM)
//   - GCP: db-n1-highmem-2 (2 vCPU, 13GB RAM) or db-n1-highmem-4 (4 vCPU, 26GB RAM)
//   - Azure: MemoryOptimized (4 vCore, 20.4GB RAM)
//
// Large Production (1000+ concurrent users):
//   - AWS: db.r5.2xlarge+ (8+ vCPU, 64+ GB RAM)
//   - GCP: db-n1-highmem-8+ (8+ vCPU, 52+ GB RAM)
//   - Azure: MemoryOptimized (8+ vCore, 40.8+ GB RAM)
//
// Security and Compliance Best Practices:
//   - Use strong passwords (12+ characters, mixed case, numbers, symbols)
//   - Store passwords in secure credential management systems (AWS Secrets Manager, etc.)
//   - Enable encryption at rest and in transit for sensitive data
//   - Use VPC security groups to restrict network access to specific IP ranges
//   - Enable audit logging for compliance requirements (GDPR, HIPAA, SOX, PCI)
//   - Regularly rotate database credentials (quarterly recommended)
//   - Use IAM database authentication when available
//   - Enable deletion protection for production databases
//   - Configure automated security patching during maintenance windows
//
// Performance Optimization:
//   - Choose instance class based on workload characteristics (CPU vs memory intensive)
//   - Start with smaller instances and scale up based on monitoring
//   - Use read replicas for read-heavy workloads (reporting, analytics)
//   - Configure appropriate storage type (SSD for performance, magnetic for cost)
//   - Enable performance insights and query monitoring
//   - Set up connection pooling in your application
//   - Monitor key metrics: CPU, memory, connections, slow queries
//
// Cost Optimization Strategies:
//   - Use reserved instances for predictable workloads (up to 60% savings)
//   - Consider Aurora Serverless for variable or intermittent workloads
//   - Right-size instances based on actual usage patterns
//   - Set up automated backups with appropriate retention periods (7-35 days)
//   - Use lifecycle policies to archive old backups to cheaper storage
//   - Monitor costs with billing alerts and cost allocation tags
//   - Consider multi-AZ only for production (adds ~100% cost)
//
// Backup and Recovery Planning:
//   - Automated backups: 7-35 day retention, point-in-time recovery
//   - Manual snapshots: Long-term retention, cross-region copies
//   - Test restore procedures regularly (monthly recommended)
//   - Document Recovery Time Objective (RTO) and Recovery Point Objective (RPO)
//   - Consider cross-region backup replication for disaster recovery
//
// Example:
//
//	config := &DBConfig{
//	    Name:              "myapp-prod-db",
//	    Engine:            "postgres",
//	    EngineVersion:     "14.9",
//	    InstanceClass:     "db.r5.large",
//	    AllocatedStorage:  100,
//	    StorageType:       "gp2",
//	    StorageEncrypted:  true,
//	    MasterUsername:    "dbadmin",
//	    MasterPassword:    "SecurePassword123!",
//	    DBName:            "myapp",
//	    VpcSecurityGroups: []string{"sg-database"},
//	    SubnetGroupName:   "db-subnet-group",
//	    MultiAZ:           true,
//	    BackupRetentionPeriod: 30,
//	    BackupWindow:      "03:00-04:00",
//	    MaintenanceWindow: "sun:04:00-sun:05:00",
//	    DeletionProtection: true,
//	    Tags: map[string]string{
//	        "Environment": "production",
//	        "Team":        "backend",
//	        "Project":     "web-app",
//	    },
//	}
type DBConfig struct {
	// Name is the unique identifier for the database instance.
	// Must be unique within your account and region
	// Length: 1-63 characters, alphanumeric and hyphens only
	// Cannot start or end with a hyphen, must start with a letter
	//
	// Naming Best Practices:
	//   - Include environment: "myapp-prod-db", "myapp-staging-db"
	//   - Include purpose: "myapp-primary-db", "myapp-analytics-db"
	//   - Use consistent naming across organization
	//   - Avoid generic names like "database" or "db"
	//
	// Examples: "ecommerce-prod-primary", "analytics-staging-replica", "cms-dev-db"
	Name string `json:"name" yaml:"name" validate:"required,min=1,max=63,hostname"`

	// Engine specifies the database engine to use.
	// Choose based on application requirements, team expertise, and ecosystem
	//
	// Supported engines by provider:
	//   AWS RDS: "mysql", "postgres", "mariadb", "oracle-ee", "oracle-se2", "sqlserver-ex", "sqlserver-web", "sqlserver-se", "sqlserver-ee"
	//   Google Cloud SQL: "mysql", "postgres", "sqlserver"
	//   Azure Database: "mysql", "postgres", "mariadb"
	//
	// Engine Selection Guide:
	//   - PostgreSQL: Best for new applications, advanced SQL features, JSON support
	//   - MySQL: Popular for web apps, large ecosystem, good performance
	//   - MariaDB: MySQL-compatible with additional features and performance
	//   - Oracle: Enterprise applications, complex queries, advanced features
	//   - SQL Server: Microsoft ecosystem, .NET applications, business intelligence
	//
	// Consider factors: licensing costs, feature requirements, team expertise, ecosystem
	Engine string `json:"engine" yaml:"engine" validate:"required,oneof=mysql postgres mariadb oracle-ee oracle-se2 sqlserver-ex sqlserver-web sqlserver-se sqlserver-ee"`

	// EngineVersion specifies the database engine version.
	// Use specific versions for production to ensure consistency and predictability
	// Avoid "latest" in production environments
	//
	// Version Selection Guidelines:
	//   - Production: Use specific versions (e.g., "14.9", "8.0.35")
	//   - Development: Can use latest minor versions for testing
	//   - Check provider documentation for supported versions
	//   - Consider Long Term Support (LTS) versions for stability
	//   - Plan for regular version upgrades (quarterly/annually)
	//
	// Current Recommended Versions:
	//   - PostgreSQL: "14.9", "15.4", "16.1" (latest stable)
	//   - MySQL: "8.0.35", "8.1.0" (latest)
	//   - MariaDB: "10.6.15", "10.11.5" (LTS)
	//   - Oracle: "19.0.0.0.ru-2023-10.rur-2023-10.r1" (19c)
	//   - SQL Server: "15.00.4335.1.v1" (2019), "16.00.4085.2.v1" (2022)
	EngineVersion string `json:"engine_version" yaml:"engine_version" validate:"required"`

	// InstanceClass defines the compute and memory capacity for the database.
	// Choose based on workload characteristics and performance requirements
	// Start small and scale up based on monitoring and performance testing
	//
	// AWS RDS Instance Classes:
	//   Burstable (T3/T4g): db.t3.micro, db.t3.small, db.t3.medium, db.t3.large
	//     - Use for: Development, testing, low-traffic applications
	//     - Features: Burstable CPU, cost-effective
	//   General Purpose (M5/M6i): db.m5.large, db.m5.xlarge, db.m6i.large
	//     - Use for: Balanced workloads, web applications
	//     - Features: Balanced CPU, memory, and network
	//   Memory Optimized (R5/R6i): db.r5.large, db.r5.xlarge, db.r6i.large
	//     - Use for: Memory-intensive applications, in-memory databases
	//     - Features: High memory-to-CPU ratio
	//
	// GCP Cloud SQL Machine Types:
	//   Shared Core: db-f1-micro, db-g1-small
	//     - Use for: Development, testing, very light workloads
	//   Standard: db-n1-standard-1, db-n1-standard-2, db-n1-standard-4
	//     - Use for: General purpose applications
	//   High Memory: db-n1-highmem-2, db-n1-highmem-4, db-n1-highmem-8
	//     - Use for: Memory-intensive applications
	//
	// Azure Database Pricing Tiers:
	//   Basic: 1-2 vCores, up to 2GB RAM
	//     - Use for: Development, testing, light workloads
	//   GeneralPurpose: 2-80 vCores, 10.2GB-408GB RAM
	//     - Use for: Most production workloads
	//   MemoryOptimized: 2-32 vCores, 20.4GB-408GB RAM
	//     - Use for: Memory-intensive applications
	InstanceClass string `json:"instance_class" yaml:"instance_class" validate:"required"`

	// AllocatedStorage is the initial storage size in gigabytes.
	// Can be increased later but not decreased (plan for growth)
	// Minimum varies by engine, typically 20GB for most engines
	//
	// Storage Planning Guidelines:
	//   - Calculate current data size + indexes + logs
	//   - Add 50-100% buffer for growth and temporary operations
	//   - Consider backup storage (not included in allocated storage)
	//   - Plan for peak usage scenarios (batch jobs, data imports)
	//
	// Minimum Storage by Engine:
	//   - MySQL: 20GB (General Purpose SSD), 100GB (Provisioned IOPS)
	//   - PostgreSQL: 20GB (General Purpose SSD), 100GB (Provisioned IOPS)
	//   - MariaDB: 20GB (General Purpose SSD), 100GB (Provisioned IOPS)
	//   - Oracle: 20GB (General Purpose SSD), 100GB (Provisioned IOPS)
	//   - SQL Server: 20GB (General Purpose SSD), 100GB (Provisioned IOPS)
	//
	// Storage Growth Considerations:
	//   - Automatic scaling available (set max storage limit)
	//   - Monitor free storage space (alert at 85% usage)
	//   - Consider data archiving strategies for historical data
	AllocatedStorage int32 `json:"allocated_storage" yaml:"allocated_storage" validate:"required,min=20,max=65536"`

	// StorageType specifies the storage technology for the database.
	// Different types offer different performance and cost characteristics
	//
	// AWS Storage Types:
	//   - "gp2": General Purpose SSD (default) - 3 IOPS/GB, burstable to 3000 IOPS
	//   - "gp3": General Purpose SSD (newer) - 3000 IOPS baseline, configurable
	//   - "io1": Provisioned IOPS SSD - Up to 64,000 IOPS, consistent performance
	//   - "io2": Provisioned IOPS SSD (newer) - Up to 256,000 IOPS, higher durability
	//   - "standard": Magnetic storage - Legacy, not recommended for production
	//
	// GCP Storage Types:
	//   - "PD_SSD": SSD persistent disks - High performance
	//   - "PD_HDD": HDD persistent disks - Cost-effective for large datasets
	//
	// Azure Storage Types:
	//   - "Premium_LRS": Premium SSD - High performance, low latency
	//   - "StandardSSD_LRS": Standard SSD - Balanced performance and cost
	//   - "Standard_LRS": Standard HDD - Cost-effective for infrequent access
	//
	// Selection Guidelines:
	//   - gp2/gp3: Most workloads, good balance of performance and cost
	//   - io1/io2: High-performance applications, consistent IOPS requirements
	//   - standard/HDD: Development, testing, archival (not recommended for production)
	StorageType string `json:"storage_type,omitempty" yaml:"storage_type,omitempty" validate:"omitempty,oneof=gp2 gp3 io1 io2 standard PD_SSD PD_HDD Premium_LRS StandardSSD_LRS Standard_LRS"`

	// StorageEncrypted enables encryption at rest for the database storage.
	// Highly recommended for production databases and required for compliance
	//
	// Benefits:
	//   - Protects data at rest from unauthorized access
	//   - Required for many compliance standards (HIPAA, PCI DSS, GDPR)
	//   - Minimal performance impact with modern hardware
	//   - Automatic key management by cloud provider
	//
	// Considerations:
	//   - Cannot be disabled after database creation
	//   - Slight increase in storage costs
	//   - Backup and snapshots are also encrypted
	//   - Read replicas inherit encryption settings
	//
	// Default: false (not encrypted) - should be true for production
	StorageEncrypted *bool `json:"storage_encrypted,omitempty" yaml:"storage_encrypted,omitempty"`

	// KmsKeyId specifies the customer-managed key for encryption.
	// Leave empty to use provider-managed keys (recommended for most use cases)
	//
	// Format varies by provider:
	//   - AWS: Key ID, ARN, or alias (e.g., "alias/my-db-key")
	//   - GCP: projects/PROJECT/locations/LOCATION/keyRings/RING/cryptoKeys/KEY
	//   - Azure: Key vault key identifier
	//
	// Customer-Managed Key Benefits:
	//   - Full control over key lifecycle
	//   - Ability to disable/delete keys
	//   - Detailed audit logs for key usage
	//   - Cross-account key sharing
	//
	// Considerations:
	//   - Additional complexity and cost
	//   - Must manage key permissions and lifecycle
	//   - Risk of data loss if key is deleted
	//   - Required for some compliance scenarios
	KmsKeyId string `json:"kms_key_id,omitempty" yaml:"kms_key_id,omitempty"`

	// MasterUsername is the primary database administrator username.
	// Cannot be reserved words or system usernames
	// Length: 1-63 characters, alphanumeric and underscores only, must start with letter
	//
	// Reserved Names to Avoid:
	//   - MySQL: "admin", "root", "mysql", "sys", "information_schema"
	//   - PostgreSQL: "postgres", "root", "admin", "rds_superuser"
	//   - SQL Server: "sa", "admin", "administrator", "root"
	//   - Oracle: "sys", "system", "admin", "root"
	//
	// Best Practices:
	//   - Use descriptive names: "dbadmin", "app_admin", "postgres_admin"
	//   - Avoid generic names like "user" or "admin"
	//   - Document the username in your team's credential management system
	//   - Consider using different usernames for different environments
	//
	// Examples: "dbadmin", "app_user", "postgres_admin", "mysql_admin"
	MasterUsername string `json:"master_username" yaml:"master_username" validate:"required,min=1,max=63,alphanum_underscore,starts_with_letter"`

	// MasterPassword is the password for the master database user.
	// Must meet complexity requirements for security and compliance
	//
	// Password Requirements:
	//   - Minimum length: 8 characters (12+ strongly recommended)
	//   - Must contain: uppercase letters, lowercase letters, numbers
	//   - Special characters recommended: !@#$%^&*()_+-=[]{}|;:,.<>?
	//   - Cannot contain: username, "password", common dictionary words
	//   - Cannot contain: quotes, backslashes, forward slashes
	//
	// Security Best Practices:
	//   - Use password managers to generate strong passwords
	//   - Store passwords in secure credential management systems
	//   - Rotate passwords regularly (quarterly recommended)
	//   - Use different passwords for different environments
	//   - Consider using IAM database authentication when available
	//   - Never log or display passwords in plain text
	//
	// Example Strong Passwords:
	//   - "MyApp2024!SecureDB#"
	//   - "Prod$Database*2024"
	//   - "Secure!DB#Pass2024"
	MasterPassword string `json:"master_password" yaml:"master_password" validate:"required,min=8,max=128,strong_password"`

	// DBName is the name of the initial database to create within the instance.
	// Leave empty to skip creating an initial database (can create later)
	// Must follow database engine naming conventions
	//
	// Naming Rules by Engine:
	//   - MySQL: 1-64 characters, alphanumeric and underscores, cannot start with number
	//   - PostgreSQL: 1-63 characters, alphanumeric and underscores, case-sensitive
	//   - SQL Server: 1-128 characters, alphanumeric and underscores, cannot be reserved words
	//   - Oracle: 1-30 characters, alphanumeric and underscores, cannot be reserved words
	//
	// Reserved Names to Avoid:
	//   - MySQL: "information_schema", "mysql", "performance_schema", "sys"
	//   - PostgreSQL: "postgres", "template0", "template1"
	//   - SQL Server: "master", "model", "msdb", "tempdb"
	//   - Oracle: "system", "sysaux", "temp", "users"
	//
	// Best Practices:
	//   - Use descriptive names related to your application
	//   - Include environment in name for clarity: "myapp_prod", "myapp_staging"
	//   - Use consistent naming conventions across your organization
	//   - Avoid spaces and special characters
	//
	// Examples: "myapp", "ecommerce_prod", "analytics_db", "cms_content"
	DBName string `json:"db_name,omitempty" yaml:"db_name,omitempty" validate:"omitempty,min=1,max=64,alphanum_underscore"`

	// VpcSecurityGroups define network access rules for the database instance.
	// Security groups act as virtual firewalls controlling inbound and outbound traffic
	// Critical for database security - restrict access to only necessary sources
	//
	// Security Group Best Practices:
	//   - Create dedicated security groups for databases (don't reuse web server groups)
	//   - Use descriptive names: "sg-myapp-database", "sg-postgres-prod"
	//   - Allow access only from application servers and admin hosts
	//   - Use specific port numbers (5432 for PostgreSQL, 3306 for MySQL)
	//   - Avoid 0.0.0.0/0 (open to internet) - major security risk
	//   - Use security group references instead of IP addresses when possible
	//   - Regularly audit and review security group rules
	//
	// Common Port Numbers:
	//   - PostgreSQL: 5432
	//   - MySQL/MariaDB: 3306
	//   - SQL Server: 1433
	//   - Oracle: 1521
	//
	// Example Security Group Rules:
	//   - Allow port 5432 from application server security group
	//   - Allow port 5432 from bastion host security group
	//   - Allow port 5432 from VPN subnet (10.0.100.0/24)
	//   - Deny all other inbound traffic
	//
	// Examples: ["sg-app-database", "sg-admin-access"], ["sg-postgres-prod", "sg-bastion"]
	VpcSecurityGroups []string `json:"vpc_security_groups,omitempty" yaml:"vpc_security_groups,omitempty"`

	// SubnetGroupName specifies the DB subnet group for the database instance.
	// Subnet groups define which subnets the database can be placed in
	// Must span multiple Availability Zones for Multi-AZ deployments
	//
	// Subnet Group Requirements:
	//   - Must contain subnets in at least two Availability Zones
	//   - All subnets must be in the same VPC
	//   - Subnets should be private (no internet gateway) for security
	//   - Must have sufficient IP addresses available
	//
	// Best Practices:
	//   - Use private subnets for database instances
	//   - Create dedicated subnet groups for databases
	//   - Use descriptive names: "db-subnet-group-prod", "postgres-subnets"
	//   - Ensure subnets have appropriate routing for application access
	//   - Consider subnet CIDR blocks for IP address planning
	//
	// Leave empty to use the default subnet group (not recommended for production)
	SubnetGroupName string `json:"subnet_group_name,omitempty" yaml:"subnet_group_name,omitempty"`

	// MultiAZ enables Multi-Availability Zone deployment for high availability.
	// Creates a standby replica in a different AZ for automatic failover
	// Recommended for production databases requiring high availability
	//
	// Benefits:
	//   - Automatic failover in case of AZ failure (typically 1-2 minutes)
	//   - Enhanced durability with synchronous replication
	//   - Automated backups from standby (no performance impact on primary)
	//   - Maintenance operations performed on standby first
	//
	// Considerations:
	//   - Approximately doubles the cost (standby instance + storage)
	//   - Slight increase in write latency due to synchronous replication
	//   - Standby is not accessible for read operations (use read replicas instead)
	//   - Required for production SLA commitments
	//
	// Default: false (single AZ) - should be true for production
	MultiAZ *bool `json:"multi_az,omitempty" yaml:"multi_az,omitempty"`

	// BackupRetentionPeriod specifies how long to retain automated backups (in days).
	// Range: 0-35 days (0 disables automated backups)
	// Longer retention provides more recovery options but increases storage costs
	//
	// Retention Guidelines:
	//   - Development: 1-7 days (minimal retention)
	//   - Staging: 7-14 days (moderate retention for testing)
	//   - Production: 14-35 days (extended retention for compliance)
	//   - Compliance: 35 days (maximum retention for regulatory requirements)
	//
	// Considerations:
	//   - Backup storage costs increase with longer retention
	//   - Point-in-time recovery available within retention period
	//   - Manual snapshots not affected by this setting
	//   - Cross-region backup copies may be required for disaster recovery
	//
	// Default: 7 days (if not specified)
	BackupRetentionPeriod *int32 `json:"backup_retention_period,omitempty" yaml:"backup_retention_period,omitempty" validate:"omitempty,min=0,max=35"`

	// BackupWindow specifies the daily time range for automated backups (UTC).
	// Format: "HH:MM-HH:MM" (must be at least 30 minutes)
	// Choose a time when database activity is lowest to minimize performance impact
	//
	// Best Practices:
	//   - Schedule during low-traffic hours (typically 2-6 AM local time)
	//   - Avoid peak business hours and batch processing times
	//   - Consider time zone differences for global applications
	//   - Coordinate with maintenance windows (should not overlap)
	//   - Allow sufficient time for backup completion
	//
	// Examples:
	//   - "03:00-04:00" (3-4 AM UTC, good for US East Coast)
	//   - "07:00-08:00" (7-8 AM UTC, good for Europe)
	//   - "14:00-15:00" (2-3 PM UTC, good for Asia Pacific)
	//
	// Leave empty to let provider choose optimal time
	BackupWindow string `json:"backup_window,omitempty" yaml:"backup_window,omitempty" validate:"omitempty,backup_window_format"`

	// MaintenanceWindow specifies the weekly time range for system maintenance (UTC).
	// Format: "ddd:HH:MM-ddd:HH:MM" (must be at least 30 minutes)
	// System updates, patches, and minor version upgrades occur during this window
	//
	// Day Abbreviations: sun, mon, tue, wed, thu, fri, sat
	//
	// Best Practices:
	//   - Schedule during lowest traffic periods
	//   - Avoid critical business hours and peak usage times
	//   - Allow sufficient time for maintenance operations (2+ hours recommended)
	//   - Coordinate with application deployment schedules
	//   - Consider impact on Multi-AZ failover during maintenance
	//
	// Examples:
	//   - "sun:03:00-sun:05:00" (Sunday 3-5 AM UTC)
	//   - "sat:06:00-sat:08:00" (Saturday 6-8 AM UTC)
	//   - "tue:02:00-tue:04:00" (Tuesday 2-4 AM UTC)
	//
	// Leave empty to let provider choose optimal time
	MaintenanceWindow string `json:"maintenance_window,omitempty" yaml:"maintenance_window,omitempty" validate:"omitempty,maintenance_window_format"`

	// DeletionProtection prevents accidental deletion of the database instance.
	// When enabled, the database cannot be deleted until protection is disabled
	// Strongly recommended for production databases
	//
	// Benefits:
	//   - Prevents accidental deletion through console, CLI, or API
	//   - Requires explicit action to disable protection before deletion
	//   - Provides additional safety for critical databases
	//   - No performance impact or additional cost
	//
	// Considerations:
	//   - Must be disabled before database can be deleted
	//   - Does not prevent data deletion within the database
	//   - Does not prevent stopping or modifying the instance
	//   - Should be part of production database checklist
	//
	// Default: false (no protection) - should be true for production
	DeletionProtection *bool `json:"deletion_protection,omitempty" yaml:"deletion_protection,omitempty"`

	// Tags are key-value pairs for organizing and managing database instances.
	// Used for billing, automation, access control, and resource organization
	// Maximum 50 tags per resource, each key/value max 255 characters
	//
	// Common Tag Patterns:
	//   - Environment: "production", "staging", "development", "testing"
	//   - Team: "backend", "frontend", "devops", "data", "platform"
	//   - Project: "web-app", "mobile-api", "data-pipeline", "analytics"
	//   - Owner: "john.doe@company.com", "team-backend@company.com"
	//   - CostCenter: "engineering", "marketing", "sales", "operations"
	//   - Application: "ecommerce", "cms", "crm", "billing"
	//   - DataClassification: "public", "internal", "confidential", "restricted"
	//   - Backup: "daily", "weekly", "monthly", "none"
	//   - Compliance: "gdpr", "hipaa", "sox", "pci", "none"
	//   - MaintenanceWindow: "weekend", "weeknight", "business-hours"
	//
	// Best Practices:
	//   - Use consistent naming conventions across organization
	//   - Include mandatory tags (Environment, Owner, Project, CostCenter)
	//   - Use tags for automated cost allocation and chargeback
	//   - Implement tag compliance policies and automation
	//   - Include data classification for security and compliance
	//   - Use tags for automated backup and maintenance scheduling
	//   - Regular audit and cleanup of unused or incorrect tags
	Tags map[string]string `json:"tags,omitempty" yaml:"tags,omitempty" validate:"max=50,dive,keys,max=255,endkeys,max=255"`

	// PerformanceInsightsEnabled enables Performance Insights monitoring.
	// Provides detailed database performance monitoring and query analysis
	// Helps identify performance bottlenecks and optimize database queries
	//
	// Benefits:
	//   - Real-time and historical performance monitoring
	//   - Query-level performance analysis and recommendations
	//   - Wait event analysis and resource utilization metrics
	//   - Integration with CloudWatch and other monitoring tools
	//   - Helps optimize database performance and reduce costs
	//
	// Considerations:
	//   - Additional cost for extended data retention (7 days free, 731 days paid)
	//   - Minimal performance overhead (< 1%)
	//   - Not available for all instance classes and engines
	//   - Data retention policies apply
	//
	// Default: false (disabled) - recommended for production databases
	PerformanceInsightsEnabled *bool `json:"performance_insights_enabled,omitempty" yaml:"performance_insights_enabled,omitempty"`

	// MonitoringInterval specifies the interval for collecting enhanced monitoring metrics.
	// Values: 0 (disabled), 1, 5, 10, 15, 30, 60 seconds
	// Enhanced monitoring provides OS-level metrics for better troubleshooting
	//
	// Monitoring Levels:
	//   - 0: Basic monitoring only (5-minute intervals)
	//   - 60: Standard enhanced monitoring (good for most production workloads)
	//   - 30: Detailed monitoring (for performance-critical applications)
	//   - 1-15: High-frequency monitoring (for troubleshooting specific issues)
	//
	// Metrics Collected:
	//   - CPU utilization by process
	//   - Memory usage and swap activity
	//   - Disk I/O and network activity
	//   - Process and thread information
	//   - File system usage
	//
	// Considerations:
	//   - Additional cost for enhanced monitoring
	//   - More frequent collection increases monitoring costs
	//   - Useful for performance troubleshooting and capacity planning
	//   - Requires IAM role for CloudWatch access
	//
	// Default: 0 (disabled)
	MonitoringInterval *int32 `json:"monitoring_interval,omitempty" yaml:"monitoring_interval,omitempty" validate:"omitempty,oneof=0 1 5 10 15 30 60"`

	// MonitoringRoleArn specifies the IAM role for enhanced monitoring.
	// Required when MonitoringInterval > 0
	// Role must have permissions to publish metrics to CloudWatch
	//
	// Format: "arn:aws:iam::account-id:role/role-name"
	// Example: "arn:aws:iam::123456789012:role/rds-monitoring-role"
	//
	// Required Permissions:
	//   - logs:CreateLogGroup
	//   - logs:CreateLogStream
	//   - logs:PutLogEvents
	//   - logs:DescribeLogStreams
	//
	// Leave empty if MonitoringInterval is 0
	MonitoringRoleArn string `json:"monitoring_role_arn,omitempty" yaml:"monitoring_role_arn,omitempty"`

	// EnabledCloudwatchLogsExports specifies which log types to export to CloudWatch.
	// Available log types vary by database engine
	// Useful for centralized logging, monitoring, and compliance
	//
	// PostgreSQL Log Types:
	//   - "postgresql": General database logs
	//   - "upgrade": Database upgrade logs
	//
	// MySQL Log Types:
	//   - "error": Error logs
	//   - "general": General query logs (high volume, use carefully)
	//   - "slow-query": Slow query logs (recommended for performance tuning)
	//
	// SQL Server Log Types:
	//   - "error": Error logs
	//   - "agent": SQL Server Agent logs
	//
	// Oracle Log Types:
	//   - "alert": Alert logs
	//   - "audit": Audit logs
	//   - "trace": Trace logs
	//   - "listener": Listener logs
	//
	// Considerations:
	//   - Additional CloudWatch Logs costs apply
	//   - General/query logs can generate high volume
	//   - Useful for security auditing and compliance
	//   - Can impact database performance if over-used
	//
	// Examples: ["postgresql"], ["error", "slow-query"], ["alert", "audit"]
	EnabledCloudwatchLogsExports []string `json:"enabled_cloudwatch_logs_exports,omitempty" yaml:"enabled_cloudwatch_logs_exports,omitempty"`
}

// DBInstance represents a managed database instance with its current state and connection information.
// This provides all the information needed to connect to and manage the database.
type DBInstance struct {
	// ID is the unique identifier assigned by the cloud provider.
	// Format varies by provider:
	//   AWS: "myapp-prod-db" (same as Name for RDS)
	//   GCP: "projects/PROJECT/instances/INSTANCE"
	//   Azure: "/subscriptions/.../resourceGroups/.../providers/Microsoft.DBforPostgreSQL/servers/server-name"
	ID string

	// Name is the human-readable name of the database instance.
	// This is the name you specified when creating the instance.
	Name string

	// Engine indicates the database engine and version.
	// Examples: "postgres-14.9", "mysql-8.0.35", "sqlserver-2019"
	// Use this to determine connection parameters and SQL dialect.
	Engine string

	// Status represents the current operational state of the database.
	// Common states:
	//   - "creating": Database is being provisioned
	//   - "available": Database is ready for connections
	//   - "modifying": Configuration changes are being applied
	//   - "backing-up": Automated backup is in progress
	//   - "maintenance": Maintenance window is active
	//   - "rebooting": Database is restarting
	//   - "deleting": Database is being deleted
	//   - "failed": Database creation or operation failed
	Status string

	// Endpoint is the connection string for accessing the database.
	// Format: hostname:port
	// Examples:
	//   - "mydb.cluster-xyz.us-east-1.rds.amazonaws.com:5432"
	//   - "10.0.1.100:3306"
	//   - "mydb.postgres.database.azure.com:5432"
	// Use this with your database driver to establish connections.
	Endpoint string

	// LaunchTime indicates when the database instance was created.
	// Format: RFC3339 timestamp (e.g., "2023-01-15T10:30:00Z")
	// Useful for tracking instance age and maintenance schedules.
	LaunchTime string
}

// Database provides managed database operations across cloud providers.
// This interface abstracts the differences between AWS RDS, Google Cloud SQL,
// Azure Database, and other managed database services.
//
// Connection String Formats by Engine:
//
// PostgreSQL:
//
//	postgres://username:password@endpoint/database?sslmode=require
//	Example: postgres://dbadmin:SecurePass123@mydb.xyz.rds.amazonaws.com:5432/myapp?sslmode=require
//
// MySQL:
//
//	mysql://username:password@tcp(endpoint)/database?tls=true
//	Example: mysql://dbadmin:SecurePass123@tcp(mydb.xyz.rds.amazonaws.com:3306)/myapp?tls=true
//
// SQL Server:
//
//	sqlserver://username:password@endpoint?database=database&encrypt=true
//	Example: sqlserver://dbadmin:SecurePass123@mydb.xyz.database.windows.net:1433?database=myapp&encrypt=true
//
// Security Best Practices:
//   - Always use SSL/TLS connections (sslmode=require, tls=true, encrypt=true)
//   - Store connection strings in environment variables or secret managers
//   - Use connection pooling to manage database connections efficiently
//   - Implement proper error handling and retry logic
//   - Monitor connection counts and query performance
//   - Use read replicas for read-heavy workloads
//
// Backup and Recovery:
//   - Automated backups are enabled by default with 7-day retention
//   - Point-in-time recovery is available for most engines
//   - Test your backup and recovery procedures regularly
//   - Consider cross-region backups for disaster recovery
//   - Document your recovery time objectives (RTO) and recovery point objectives (RPO)
//
// Monitoring and Maintenance:
//   - Enable performance insights and query monitoring
//   - Set up alerts for CPU, memory, and storage usage
//   - Monitor connection counts and slow queries
//   - Schedule maintenance windows during low-traffic periods
//   - Keep database engines updated with security patches
type Database interface {
	// CreateDB creates a new managed database instance with the specified configuration.
	// The database will be created asynchronously - use GetDB to monitor the creation progress.
	// Initial creation typically takes 10-20 minutes depending on the engine and instance size.
	//
	// Configuration Validation:
	//   - Name: Must be unique within your account and region
	//   - Engine: Must be supported by the provider in the target region
	//   - EngineVersion: Must be a valid version for the chosen engine
	//   - InstanceClass: Must be available in the target region
	//   - AllocatedStorage: Must meet minimum requirements for the engine
	//   - MasterUsername: Cannot be reserved words (admin, root, etc.)
	//   - MasterPassword: Must meet complexity requirements
	//   - VpcSecurityGroups: All groups must exist in the target VPC
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials or expired tokens
	//   - ErrAuthorization: Insufficient permissions or database limits exceeded
	//   - ErrInvalidConfig: Invalid database configuration or unsupported options
	//   - ErrResourceConflict: Database name already exists
	//   - ErrResourceNotFound: Security groups or subnet groups don't exist
	//   - ErrRateLimit: Too many database creation requests
	//
	// Example:
	//   config := &DBConfig{
	//       Name:              "myapp-prod-db",
	//       Engine:            "postgres",
	//       EngineVersion:     "14.9",
	//       InstanceClass:     "db.t3.micro",
	//       AllocatedStorage:  20,
	//       MasterUsername:    "dbadmin",
	//       MasterPassword:    "SecurePassword123!",
	//       DBName:            "myapp",
	//       VpcSecurityGroups: []string{"sg-database"},
	//   }
	//
	//   db, err := database.CreateDB(ctx, config)
	//   if err != nil {
	//       log.Fatalf("Failed to create database: %v", err)
	//   }
	//
	//   fmt.Printf("Database creation initiated: %s\n", db.Name)
	//   fmt.Printf("Status: %s\n", db.Status)
	//
	//   // Wait for database to become available
	//   fmt.Println("Waiting for database to become available...")
	//   for db.Status != "available" {
	//       time.Sleep(30 * time.Second)
	//       db, err = database.GetDB(ctx, db.ID)
	//       if err != nil {
	//           log.Fatalf("Failed to check database status: %v", err)
	//       }
	//       fmt.Printf("Current status: %s\n", db.Status)
	//   }
	//
	//   // Database is ready - construct connection string
	//   connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=require",
	//       config.MasterUsername, config.MasterPassword, db.Endpoint, config.DBName)
	//
	//   fmt.Printf("Database is ready!\n")
	//   fmt.Printf("Endpoint: %s\n", db.Endpoint)
	//   fmt.Printf("Connection string: postgres://[username]:[password]@%s/%s?sslmode=require\n",
	//       db.Endpoint, config.DBName)
	CreateDB(ctx context.Context, config *DBConfig) (*DBInstance, error)

	// ListDBs returns all database instances in the current region.
	// Returns an empty slice if no databases exist.
	// Results include databases in all states (creating, available, deleting, etc.).
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials or expired tokens
	//   - ErrAuthorization: Insufficient permissions to list database instances
	//   - ErrRateLimit: Too many list requests
	//
	// Example:
	//   databases, err := database.ListDBs(ctx)
	//   if err != nil {
	//       log.Fatalf("Failed to list databases: %v", err)
	//   }
	//
	//   if len(databases) == 0 {
	//       fmt.Println("No databases found. Create one with CreateDB().")
	//       return
	//   }
	//
	//   fmt.Printf("Found %d databases:\n", len(databases))
	//   for i, db := range databases {
	//       fmt.Printf("  %d. %s (%s)\n", i+1, db.Name, db.Engine)
	//       fmt.Printf("     Status: %s\n", db.Status)
	//       if db.Status == "available" {
	//           fmt.Printf("     Endpoint: %s\n", db.Endpoint)
	//       }
	//       fmt.Printf("     Created: %s\n", db.LaunchTime)
	//       fmt.Println()
	//   }
	//
	//   // Filter by engine type
	//   postgresDBs := make([]*DBInstance, 0)
	//   for _, db := range databases {
	//       if strings.Contains(db.Engine, "postgres") {
	//           postgresDBs = append(postgresDBs, db)
	//       }
	//   }
	//   fmt.Printf("PostgreSQL databases: %d\n", len(postgresDBs))
	//
	//   // Find available databases
	//   availableDBs := make([]*DBInstance, 0)
	//   for _, db := range databases {
	//       if db.Status == "available" {
	//           availableDBs = append(availableDBs, db)
	//       }
	//   }
	//   fmt.Printf("Available databases: %d\n", len(availableDBs))
	ListDBs(ctx context.Context) ([]*DBInstance, error)

	// GetDB retrieves detailed information about a specific database instance.
	// Returns the current status, endpoint, and configuration details.
	// Use this method to monitor database status during creation, modification, or deletion.
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials or expired tokens
	//   - ErrAuthorization: Insufficient permissions or database not owned by your account
	//   - ErrResourceNotFound: Database with the specified ID doesn't exist
	//   - ErrRateLimit: Too many describe requests
	//
	// Example:
	//   db, err := database.GetDB(ctx, "myapp-prod-db")
	//   if err != nil {
	//       if errors.Is(err, ErrResourceNotFound) {
	//           fmt.Println("Database not found - it may have been deleted")
	//           return
	//       }
	//       log.Fatalf("Failed to get database details: %v", err)
	//   }
	//
	//   fmt.Printf("Database Details:\n")
	//   fmt.Printf("  Name: %s\n", db.Name)
	//   fmt.Printf("  Engine: %s\n", db.Engine)
	//   fmt.Printf("  Status: %s\n", db.Status)
	//   fmt.Printf("  Created: %s\n", db.LaunchTime)
	//
	//   if db.Status == "available" {
	//       fmt.Printf("  Endpoint: %s\n", db.Endpoint)
	//
	//       // Extract host and port for connection
	//       parts := strings.Split(db.Endpoint, ":")
	//       if len(parts) == 2 {
	//           host, port := parts[0], parts[1]
	//           fmt.Printf("  Host: %s\n", host)
	//           fmt.Printf("  Port: %s\n", port)
	//       }
	//
	//       // Show connection examples for different engines
	//       if strings.Contains(db.Engine, "postgres") {
	//           fmt.Printf("  Connection example: psql -h %s -U dbadmin -d myapp\n",
	//               strings.Split(db.Endpoint, ":")[0])
	//       } else if strings.Contains(db.Engine, "mysql") {
	//           fmt.Printf("  Connection example: mysql -h %s -u dbadmin -p myapp\n",
	//               strings.Split(db.Endpoint, ":")[0])
	//       }
	//   } else {
	//       fmt.Printf("  Database is not ready for connections (status: %s)\n", db.Status)
	//   }
	GetDB(ctx context.Context, id string) (*DBInstance, error)

	// DeleteDB permanently deletes a database instance and all its data.
	// This operation cannot be undone. All databases, tables, and data will be lost forever.
	// A final snapshot may be created automatically before deletion (provider-dependent).
	//
	// Deletion Process:
	//   1. Database stops accepting new connections
	//   2. Existing connections are terminated
	//   3. Final snapshot is created (if configured)
	//   4. Database instance is deleted
	//   5. Associated storage is released
	//
	// Important Considerations:
	//   - Deletion protection may be enabled (must be disabled first)
	//   - Automated backups are deleted after the retention period
	//   - Manual snapshots are preserved unless explicitly deleted
	//   - Read replicas must be deleted before the primary instance
	//   - Some providers require final snapshot confirmation
	//
	// Common errors:
	//   - ErrAuthentication: Invalid credentials or expired tokens
	//   - ErrAuthorization: Insufficient permissions or database not owned by your account
	//   - ErrResourceNotFound: Database with the specified ID doesn't exist
	//   - ErrResourceConflict: Database has deletion protection enabled or has read replicas
	//   - ErrInvalidConfig: Database is in a state that prevents deletion
	//   - ErrRateLimit: Too many delete requests
	//
	// Example:
	//   dbID := "old-test-database"
	//
	//   // Get database details before deletion
	//   db, err := database.GetDB(ctx, dbID)
	//   if err != nil {
	//       log.Fatalf("Failed to get database details: %v", err)
	//   }
	//
	//   // Confirm deletion (this is permanent!)
	//   fmt.Printf("WARNING: You are about to delete database '%s' (%s)\n", db.Name, db.Engine)
	//   fmt.Printf("This will permanently delete all data and cannot be undone.\n")
	//   fmt.Print("Type 'DELETE' to confirm: ")
	//
	//   var confirmation string
	//   fmt.Scanln(&confirmation)
	//   if confirmation != "DELETE" {
	//       fmt.Println("Deletion cancelled")
	//       return
	//   }
	//
	//   // Perform the deletion
	//   err = database.DeleteDB(ctx, dbID)
	//   if err != nil {
	//       if errors.Is(err, ErrResourceConflict) {
	//           fmt.Println("Cannot delete database: deletion protection may be enabled or read replicas exist")
	//           return
	//       }
	//       log.Fatalf("Failed to delete database: %v", err)
	//   }
	//
	//   fmt.Printf("Database deletion initiated: %s\n", db.Name)
	//   fmt.Println("The database will be permanently deleted shortly.")
	//
	//   // Optional: Wait for deletion to complete
	//   fmt.Println("Waiting for deletion to complete...")
	//   for {
	//       _, err := database.GetDB(ctx, dbID)
	//       if err != nil {
	//           if errors.Is(err, ErrResourceNotFound) {
	//               fmt.Println("Database has been successfully deleted")
	//               break
	//           }
	//           log.Printf("Error checking database status: %v", err)
	//           break
	//       }
	//
	//       fmt.Print(".")
	//       time.Sleep(30 * time.Second)
	//   }
	//
	//   fmt.Println("\nDeletion complete. Remember to:")
	//   fmt.Println("  - Update application connection strings")
	//   fmt.Println("  - Remove database credentials from secret managers")
	//   fmt.Println("  - Update security group rules if no longer needed")
	//   fmt.Println("  - Check for any remaining manual snapshots")
	DeleteDB(ctx context.Context, id string) error
}
