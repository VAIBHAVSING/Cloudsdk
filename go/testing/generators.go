package testing

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/VAIBHAVSING/Cloudsdk/go/services"
)

// Test data generators provide realistic configurations for testing

// GenerateVMConfig creates a realistic VM configuration for testing
func GenerateVMConfig(name string) *services.VMConfig {
	return &services.VMConfig{
		Name:           name,
		ImageID:        GenerateImageID(),
		InstanceType:   GenerateInstanceType(),
		KeyName:        "test-keypair",
		SecurityGroups: []string{"sg-test"},
		Tags: map[string]string{
			"Environment": "test",
			"Purpose":     "testing",
		},
	}
}

// GenerateVMConfigWithOptions creates a VM configuration with specific options
func GenerateVMConfigWithOptions(name string, options VMConfigOptions) *services.VMConfig {
	config := GenerateVMConfig(name)

	if options.ImageID != "" {
		config.ImageID = options.ImageID
	}
	if options.InstanceType != "" {
		config.InstanceType = options.InstanceType
	}
	if options.KeyName != "" {
		config.KeyName = options.KeyName
	}
	if len(options.SecurityGroups) > 0 {
		config.SecurityGroups = options.SecurityGroups
	}
	if options.UserData != "" {
		config.UserData = options.UserData
	}
	if len(options.Tags) > 0 {
		config.Tags = options.Tags
	}

	return config
}

// VMConfigOptions provides options for customizing generated VM configurations
type VMConfigOptions struct {
	ImageID        string
	InstanceType   string
	KeyName        string
	SecurityGroups []string
	UserData       string
	Tags           map[string]string
}

// GenerateBucketConfig creates a realistic bucket configuration for testing
func GenerateBucketConfig(name string) *services.BucketConfig {
	return &services.BucketConfig{
		Name:   name,
		Region: "us-east-1",
		Tags: map[string]string{
			"Environment": "test",
			"Purpose":     "testing",
		},
	}
}

// GenerateBucketConfigWithOptions creates a bucket configuration with specific options
func GenerateBucketConfigWithOptions(name string, options BucketConfigOptions) *services.BucketConfig {
	config := GenerateBucketConfig(name)

	if options.Region != "" {
		config.Region = options.Region
	}
	if options.Versioning != nil {
		config.Versioning = options.Versioning
	}
	if options.ACL != "" {
		config.ACL = options.ACL
	}
	if options.StorageClass != "" {
		config.StorageClass = options.StorageClass
	}
	if len(options.Tags) > 0 {
		config.Tags = options.Tags
	}

	return config
}

// BucketConfigOptions provides options for customizing generated bucket configurations
type BucketConfigOptions struct {
	Region       string
	Versioning   *bool
	ACL          string
	StorageClass string
	Tags         map[string]string
}

// GenerateDBConfig creates a realistic database configuration for testing
func GenerateDBConfig(name string) *services.DBConfig {
	return &services.DBConfig{
		Name:             name,
		Engine:           "postgres",
		EngineVersion:    "14.9",
		InstanceClass:    "db.t3.micro",
		AllocatedStorage: 20,
		MasterUsername:   "testuser",
		MasterPassword:   "TestPassword123!",
		DBName:           "testdb",
		Tags: map[string]string{
			"Environment": "test",
			"Purpose":     "testing",
		},
	}
}

// GenerateDBConfigWithOptions creates a database configuration with specific options
func GenerateDBConfigWithOptions(name string, options DBConfigOptions) *services.DBConfig {
	config := GenerateDBConfig(name)

	if options.Engine != "" {
		config.Engine = options.Engine
	}
	if options.EngineVersion != "" {
		config.EngineVersion = options.EngineVersion
	}
	if options.InstanceClass != "" {
		config.InstanceClass = options.InstanceClass
	}
	if options.AllocatedStorage > 0 {
		config.AllocatedStorage = options.AllocatedStorage
	}
	if options.MasterUsername != "" {
		config.MasterUsername = options.MasterUsername
	}
	if options.MasterPassword != "" {
		config.MasterPassword = options.MasterPassword
	}
	if options.DBName != "" {
		config.DBName = options.DBName
	}
	if options.MultiAZ != nil {
		config.MultiAZ = options.MultiAZ
	}
	if options.StorageEncrypted != nil {
		config.StorageEncrypted = options.StorageEncrypted
	}
	if len(options.Tags) > 0 {
		config.Tags = options.Tags
	}

	return config
}

// DBConfigOptions provides options for customizing generated database configurations
type DBConfigOptions struct {
	Engine           string
	EngineVersion    string
	InstanceClass    string
	AllocatedStorage int32
	MasterUsername   string
	MasterPassword   string
	DBName           string
	MultiAZ          *bool
	StorageEncrypted *bool
	Tags             map[string]string
}

// Random data generators

// GenerateImageID generates a realistic AMI ID for testing
func GenerateImageID() string {
	return fmt.Sprintf("ami-%016x", rand.Int63())
}

// GenerateInstanceType returns a random instance type for testing
func GenerateInstanceType() string {
	types := []string{
		"t2.micro", "t2.small", "t2.medium", "t2.large",
		"t3.micro", "t3.small", "t3.medium", "t3.large",
		"m5.large", "m5.xlarge", "m5.2xlarge",
		"c5.large", "c5.xlarge", "c5.2xlarge",
	}
	return types[rand.Intn(len(types))]
}

// GenerateVMID generates a realistic VM ID for testing
func GenerateVMID() string {
	return fmt.Sprintf("i-%016x", rand.Int63())
}

// GenerateDBInstanceID generates a realistic database instance ID for testing
func GenerateDBInstanceID() string {
	return fmt.Sprintf("db-%d", time.Now().Unix())
}

// GenerateEndpoint generates a realistic database endpoint for testing
func GenerateEndpoint(name, region string) string {
	return fmt.Sprintf("%s.cluster-%x.%s.rds.amazonaws.com", name, rand.Int31(), region)
}

// GenerateBucketName generates a valid bucket name for testing
func GenerateBucketName(prefix string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-test-%d", prefix, timestamp)
}

// GenerateObjectKey generates a realistic object key for testing
func GenerateObjectKey() string {
	paths := []string{
		"documents/file.pdf",
		"images/photo.jpg",
		"data/export.csv",
		"logs/application.log",
		"backups/database.sql",
	}
	return paths[rand.Intn(len(paths))]
}

// GenerateSecurityGroupID generates a realistic security group ID for testing
func GenerateSecurityGroupID() string {
	return fmt.Sprintf("sg-%016x", rand.Int63())
}

// GenerateSubnetID generates a realistic subnet ID for testing
func GenerateSubnetID() string {
	return fmt.Sprintf("subnet-%016x", rand.Int63())
}

// GenerateVPCID generates a realistic VPC ID for testing
func GenerateVPCID() string {
	return fmt.Sprintf("vpc-%016x", rand.Int63())
}

// GenerateKeyPairName generates a realistic key pair name for testing
func GenerateKeyPairName() string {
	return fmt.Sprintf("keypair-%d", time.Now().Unix())
}

// GenerateUserData generates realistic user data for testing
func GenerateUserData() string {
	scripts := []string{
		"#!/bin/bash\nyum update -y\nyum install -y httpd\nsystemctl start httpd",
		"#!/bin/bash\napt-get update\napt-get install -y nginx\nsystemctl start nginx",
		"#!/bin/bash\necho 'Hello, World!' > /var/www/html/index.html",
		"#cloud-config\npackages:\n  - docker\n  - git",
	}
	return scripts[rand.Intn(len(scripts))]
}

// GenerateTags generates realistic tags for testing
func GenerateTags() map[string]string {
	environments := []string{"development", "staging", "production", "test"}
	teams := []string{"backend", "frontend", "devops", "data"}
	projects := []string{"web-app", "mobile-api", "data-pipeline", "analytics"}

	return map[string]string{
		"Environment": environments[rand.Intn(len(environments))],
		"Team":        teams[rand.Intn(len(teams))],
		"Project":     projects[rand.Intn(len(projects))],
		"Owner":       "test@example.com",
		"CreatedBy":   "test-suite",
	}
}

// Batch generators for load testing

// GenerateVMConfigs generates multiple VM configurations for batch testing
func GenerateVMConfigs(count int, namePrefix string) []*services.VMConfig {
	configs := make([]*services.VMConfig, count)
	for i := 0; i < count; i++ {
		configs[i] = GenerateVMConfig(fmt.Sprintf("%s-%d", namePrefix, i))
	}
	return configs
}

// GenerateBucketConfigs generates multiple bucket configurations for batch testing
func GenerateBucketConfigs(count int, namePrefix string) []*services.BucketConfig {
	configs := make([]*services.BucketConfig, count)
	for i := 0; i < count; i++ {
		configs[i] = GenerateBucketConfig(fmt.Sprintf("%s-%d", namePrefix, i))
	}
	return configs
}

// GenerateDBConfigs generates multiple database configurations for batch testing
func GenerateDBConfigs(count int, namePrefix string) []*services.DBConfig {
	configs := make([]*services.DBConfig, count)
	for i := 0; i < count; i++ {
		configs[i] = GenerateDBConfig(fmt.Sprintf("%s-%d", namePrefix, i))
	}
	return configs
}

// Realistic data patterns

// GenerateProductionVMConfig generates a production-like VM configuration
func GenerateProductionVMConfig(name string) *services.VMConfig {
	return &services.VMConfig{
		Name:           name,
		ImageID:        "ami-0abcdef1234567890", // Amazon Linux 2
		InstanceType:   "m5.large",
		KeyName:        "production-keypair",
		SecurityGroups: []string{"sg-web", "sg-ssh"},
		UserData:       "#!/bin/bash\nyum update -y\nyum install -y httpd\nsystemctl start httpd\nsystemctl enable httpd",
		Tags: map[string]string{
			"Environment": "production",
			"Team":        "backend",
			"Project":     "web-app",
			"Backup":      "daily",
			"Monitoring":  "enabled",
		},
	}
}

// GenerateProductionDBConfig generates a production-like database configuration
func GenerateProductionDBConfig(name string) *services.DBConfig {
	multiAZ := true
	encrypted := true
	deletionProtection := true
	backupRetention := int32(30)

	return &services.DBConfig{
		Name:                  name,
		Engine:                "postgres",
		EngineVersion:         "14.9",
		InstanceClass:         "db.r5.large",
		AllocatedStorage:      100,
		StorageType:           "gp2",
		StorageEncrypted:      &encrypted,
		MasterUsername:        "dbadmin",
		MasterPassword:        "SecureProductionPassword123!",
		DBName:                "production",
		MultiAZ:               &multiAZ,
		BackupRetentionPeriod: &backupRetention,
		BackupWindow:          "03:00-04:00",
		MaintenanceWindow:     "sun:04:00-sun:06:00",
		DeletionProtection:    &deletionProtection,
		VpcSecurityGroups:     []string{"sg-database"},
		Tags: map[string]string{
			"Environment": "production",
			"Team":        "backend",
			"Project":     "web-app",
			"Backup":      "daily",
			"Compliance":  "required",
		},
	}
}

// GenerateProductionBucketConfig generates a production-like bucket configuration
func GenerateProductionBucketConfig(name string) *services.BucketConfig {
	versioning := true

	return &services.BucketConfig{
		Name:         name,
		Region:       "us-east-1",
		Versioning:   &versioning,
		ACL:          "private",
		StorageClass: "STANDARD",
		Tags: map[string]string{
			"Environment": "production",
			"Team":        "backend",
			"Project":     "web-app",
			"DataClass":   "confidential",
			"Backup":      "enabled",
		},
	}
}
