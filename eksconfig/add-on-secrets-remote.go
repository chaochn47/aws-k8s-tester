package eksconfig

import (
	"errors"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-k8s-tester/pkg/metrics"
	"github.com/aws/aws-k8s-tester/pkg/randutil"
	"github.com/aws/aws-k8s-tester/pkg/timeutil"
)

// AddOnSecretsRemote defines parameters for EKS cluster
// add-on "Secrets" remote.
// It generates loads from the remote workers (Pod) in the cluster.
// Each worker writes serially with no concurrency.
// Configure "DeploymentReplicas" accordingly to increase the concurrency.
// The main use case is to write a large number of objects to fill up etcd database.
// And measure latencies for secret encryption.
type AddOnSecretsRemote struct {
	// Enable is 'true' to create this add-on.
	Enable bool `json:"enable"`
	// Created is true when the resource has been created.
	// Used for delete operations.
	Created         bool               `json:"created" read-only:"true"`
	TimeFrameCreate timeutil.TimeFrame `json:"time-frame-create" read-only:"true"`
	TimeFrameDelete timeutil.TimeFrame `json:"time-frame-delete" read-only:"true"`

	// Namespace is the namespace to create objects in.
	Namespace string `json:"namespace"`

	// RepositoryAccountID is the account ID for tester ECR image.
	// e.g. "aws/aws-k8s-tester" for "[ACCOUNT_ID].dkr.ecr.[REGION].amazonaws.com/aws/aws-k8s-tester"
	RepositoryAccountID string `json:"repository-account-id,omitempty"`
	// RepositoryName is the repositoryName for tester ECR image.
	// e.g. "aws/aws-k8s-tester" for "[ACCOUNT_ID].dkr.ecr.[REGION].amazonaws.com/aws/aws-k8s-tester"
	RepositoryName string `json:"repository-name,omitempty"`
	// RepositoryImageTag is the image tag for tester ECR image.
	// e.g. "latest" for image URI "[ACCOUNT_ID].dkr.ecr.[REGION].amazonaws.com/aws/aws-k8s-tester:latest"
	RepositoryImageTag string `json:"repository-image-tag,omitempty"`

	// DeploymentReplicas is the number of replicas to create for workers.
	// The total number of objects to be created is "DeploymentReplicas" * "Objects".
	DeploymentReplicas int32 `json:"deployment-replicas,omitempty"`
	// Objects is the number of "Secret" objects to write/read.
	Objects int `json:"objects"`
	// ObjectSize is the "Secret" value size in bytes.
	ObjectSize int `json:"object-size"`

	// NamePrefix is the prefix of Secret name.
	// If multiple Secret loader is running,
	// this must be unique per worker to avoid name conflicts.
	NamePrefix string `json:"name-prefix"`

	// S3Dir is the S3 directory to store all test results.
	// It is under the bucket "eksconfig.Config.S3BucketName".
	S3Dir string `json:"s3-dir"`

	// RequestsWritesRawJSONPath is the file path to store writes requests in JSON format.
	RequestsWritesRawJSONPath  string `json:"requests-writes-json-path" read-only:"true"`
	RequestsWritesRawJSONS3Key string `json:"requests-writes-json-s3-key" read-only:"true"`
	// RequestsWritesSummary is the writes results.
	RequestsWritesSummary metrics.RequestsSummary `json:"requests-writes-summary,omitempty" read-only:"true"`
	// RequestsWritesSummaryJSONPath is the file path to store writes requests summary in JSON format.
	RequestsWritesSummaryJSONPath  string `json:"requests-writes-summary-json-path" read-only:"true"`
	RequestsWritesSummaryJSONS3Key string `json:"requests-writes-summary-json-s3-key" read-only:"true"`
	// RequestsWritesSummaryTablePath is the file path to store writes requests summary in table format.
	RequestsWritesSummaryTablePath  string `json:"requests-writes-summary-table-path" read-only:"true"`
	RequestsWritesSummaryTableS3Key string `json:"requests-writes-summary-table-s3-path" read-only:"true"`
	// RequestsWritesSummaryS3Dir is the S3 directory of previous/latest "RequestsWritesSummary".
	// Specify the S3 key in the same bucket of "eksconfig.Config.S3BucketName".
	// Use for regression tests. Specify the value not bound to the cluster directory.
	// Different runs from different clusters reads and writes in this directory.
	RequestsWritesSummaryS3Dir string `json:"requests-writes-summary-s3-dir"`
	// RequestsWritesSummaryCompare is the comparision results.
	RequestsWritesSummaryCompare metrics.RequestsSummaryCompare `json:"requests-writes-summary-compare" read-only:"true"`
	// RequestsWritesSummaryCompareJSONPath is the file path to store writes requests compare summary in JSON format.
	RequestsWritesSummaryCompareJSONPath  string `json:"requests-writes-summary-compare-json-path" read-only:"true"`
	RequestsWritesSummaryCompareJSONS3Key string `json:"requests-writes-summary-compare-json-s3-key" read-only:"true"`
	// RequestsWritesSummaryCompareTablePath is the file path to store writes requests compare summary in table format.
	RequestsWritesSummaryCompareTablePath  string `json:"requests-writes-summary-compare-table-path" read-only:"true"`
	RequestsWritesSummaryCompareTableS3Key string `json:"requests-writes-summary-compare-table-s3-path" read-only:"true"`

	// RequestsReadsRawJSONPath is the file path to store reads requests in JSON format.
	RequestsReadsRawJSONPath  string `json:"requests-reads-raw-json-path" read-only:"true"`
	RequestsReadsRawJSONS3Key string `json:"requests-reads-raw-json-s3-key" read-only:"true"`
	// RequestsReadsSummary is the reads results.
	RequestsReadsSummary metrics.RequestsSummary `json:"requests-reads-summary,omitempty" read-only:"true"`
	// RequestsReadsSummaryJSONPath is the file path to store reads requests summary in JSON format.
	RequestsReadsSummaryJSONPath  string `json:"requests-reads-summary-json-path" read-only:"true"`
	RequestsReadsSummaryJSONS3Key string `json:"requests-reads-summary-json-s3-key" read-only:"true"`
	// RequestsReadsSummaryTablePath is the file path to store reads requests summary in table format.
	RequestsReadsSummaryTablePath  string `json:"requests-reads-summary-table-path" read-only:"true"`
	RequestsReadsSummaryTableS3Key string `json:"requests-reads-summary-table-s3-path" read-only:"true"`
	// RequestsReadsSummaryS3Dir is the S3 directory of previous/latest "RequestsReadsSummary".
	// Specify the S3 key in the same bucket of "eksconfig.Config.S3BucketName".
	// Use for regression tests. Specify the value not bound to the cluster directory.
	// Different runs from different clusters reads and writes in this directory.
	RequestsReadsSummaryS3Dir string `json:"requests-reads-summary-s3-dir"`
	// RequestsReadsSummaryCompare is the comparision results.
	RequestsReadsSummaryCompare metrics.RequestsSummaryCompare `json:"requests-reads-summary-compare" read-only:"true"`
	// RequestsReadsSummaryCompareJSONPath is the file path to store reads requests compare summary in JSON format.
	RequestsReadsSummaryCompareJSONPath  string `json:"requests-reads-summary-compare-json-path" read-only:"true"`
	RequestsReadsSummaryCompareJSONS3Key string `json:"requests-reads-summary-compare-json-s3-key" read-only:"true"`
	// RequestsReadsSummaryCompareTablePath is the file path to store reads requests compare summary in table format.
	RequestsReadsSummaryCompareTablePath  string `json:"requests-reads-summary-compare-table-path" read-only:"true"`
	RequestsReadsSummaryCompareTableS3Key string `json:"requests-reads-summary-compare-table-s3-path" read-only:"true"`

	// RequestsWritesSummaryOutputNamePrefix is the output path name in "/var/log" directory, used in remote worker.
	RequestsWritesSummaryOutputNamePrefix string `json:"requests-writes-summary-output-name-prefix"`
	// RequestsReadsSummaryOutputNamePrefix is the output path name in "/var/log" directory, used in remote worker.
	RequestsReadsSummaryOutputNamePrefix string `json:"requests-reads-summary-output-name-prefix"`
}

// EnvironmentVariablePrefixAddOnSecretsRemote is the environment variable prefix used for "eksconfig".
const EnvironmentVariablePrefixAddOnSecretsRemote = AWS_K8S_TESTER_EKS_PREFIX + "ADD_ON_SECRETS_REMOTE_"

// IsEnabledAddOnSecretsRemote returns true if "AddOnSecretsRemote" is enabled.
// Otherwise, nil the field for "omitempty".
func (cfg *Config) IsEnabledAddOnSecretsRemote() bool {
	if cfg.AddOnSecretsRemote == nil {
		return false
	}
	if cfg.AddOnSecretsRemote.Enable {
		return true
	}
	cfg.AddOnSecretsRemote = nil
	return false
}

func getDefaultAddOnSecretsRemote() *AddOnSecretsRemote {
	return &AddOnSecretsRemote{
		Enable:             false,
		DeploymentReplicas: 5,
		Objects:            10,
		ObjectSize:         10 * 1024, // 10 KB

		// writes total 100 MB for "Secret" objects,
		// plus "Pod" objects, writes total 330 MB to etcd
		//
		// with 3 nodes, takes about 1.5 hour for all
		// these "Pod"s to complete
		//
		// Objects: 10000,
		// ObjectSize: 10 * 1024, // 10 KB

		NamePrefix: "secret" + randutil.String(5),

		RequestsWritesSummaryOutputNamePrefix: "secrets-writes-" + randutil.String(10),
		RequestsReadsSummaryOutputNamePrefix:  "secrets-reads-" + randutil.String(10),
	}
}

func (cfg *Config) validateAddOnSecretsRemote() error {
	if !cfg.IsEnabledAddOnSecretsRemote() {
		return nil
	}
	if cfg.S3BucketName == "" {
		return errors.New("AddOnSecretsRemote requires S3 bucket for collecting results but S3BucketName empty")
	}

	if !cfg.IsEnabledAddOnNodeGroups() && !cfg.IsEnabledAddOnManagedNodeGroups() {
		return errors.New("AddOnSecretsRemote.Enable true but no node group is enabled")
	}

	if cfg.AddOnSecretsRemote.Namespace == "" {
		cfg.AddOnSecretsRemote.Namespace = cfg.Name + "-secrets-remote"
	}

	if cfg.AddOnSecretsRemote.RepositoryAccountID == "" {
		return errors.New("AddOnSecretsRemote.RepositoryAccountID empty")
	}
	if cfg.AddOnSecretsRemote.RepositoryName == "" {
		return errors.New("AddOnSecretsRemote.RepositoryName empty")
	}
	if cfg.AddOnSecretsRemote.RepositoryImageTag == "" {
		return errors.New("AddOnSecretsRemote.RepositoryImageTag empty")
	}

	if cfg.AddOnSecretsRemote.DeploymentReplicas == 0 {
		cfg.AddOnSecretsRemote.DeploymentReplicas = 5
	}
	if cfg.AddOnSecretsRemote.Objects == 0 {
		cfg.AddOnSecretsRemote.Objects = 10
	}
	if cfg.AddOnSecretsRemote.ObjectSize == 0 {
		cfg.AddOnSecretsRemote.ObjectSize = 10 * 1024
	}

	if cfg.AddOnSecretsRemote.NamePrefix == "" {
		cfg.AddOnSecretsRemote.NamePrefix = "secret" + randutil.String(5)
	}

	if cfg.AddOnSecretsRemote.S3Dir == "" {
		cfg.AddOnSecretsRemote.S3Dir = path.Join(cfg.Name, "add-on-secrets-remote")
	}

	if cfg.AddOnSecretsRemote.RequestsWritesRawJSONPath == "" {
		cfg.AddOnSecretsRemote.RequestsWritesRawJSONPath = strings.ReplaceAll(cfg.ConfigPath, ".yaml", "") + "-secrets-remote-requests-writes-raw.json"
	}
	if cfg.AddOnSecretsRemote.RequestsWritesRawJSONS3Key == "" {
		cfg.AddOnSecretsRemote.RequestsWritesRawJSONS3Key = path.Join(
			cfg.AddOnSecretsRemote.S3Dir,
			"writes-raw",
			filepath.Base(cfg.AddOnSecretsRemote.RequestsWritesRawJSONPath),
		)
	}
	if cfg.AddOnSecretsRemote.RequestsWritesSummaryJSONPath == "" {
		cfg.AddOnSecretsRemote.RequestsWritesSummaryJSONPath = strings.ReplaceAll(cfg.ConfigPath, ".yaml", "") + "-secrets-remote-requests-writes-summary.json"
	}
	if cfg.AddOnSecretsRemote.RequestsWritesSummaryJSONS3Key == "" {
		cfg.AddOnSecretsRemote.RequestsWritesSummaryJSONS3Key = path.Join(
			cfg.AddOnSecretsRemote.S3Dir,
			"writes-summary",
			filepath.Base(cfg.AddOnSecretsRemote.RequestsWritesSummaryJSONPath),
		)
	}
	if cfg.AddOnSecretsRemote.RequestsWritesSummaryTablePath == "" {
		cfg.AddOnSecretsRemote.RequestsWritesSummaryTablePath = strings.ReplaceAll(cfg.ConfigPath, ".yaml", "") + "-secrets-remote-requests-writes-summary.txt"
	}
	if cfg.AddOnSecretsRemote.RequestsWritesSummaryTableS3Key == "" {
		cfg.AddOnSecretsRemote.RequestsWritesSummaryTableS3Key = path.Join(
			cfg.AddOnSecretsRemote.S3Dir,
			"writes-summary",
			filepath.Base(cfg.AddOnSecretsRemote.RequestsWritesSummaryTablePath),
		)
	}
	if cfg.AddOnSecretsRemote.RequestsWritesSummaryS3Dir == "" {
		cfg.AddOnSecretsRemote.RequestsWritesSummaryS3Dir = path.Join("add-on-secrets-remote", "writes-summary", cfg.Parameters.Version)
	}
	if cfg.AddOnSecretsRemote.RequestsWritesSummaryCompareJSONPath == "" {
		cfg.AddOnSecretsRemote.RequestsWritesSummaryCompareJSONPath = strings.ReplaceAll(cfg.ConfigPath, ".yaml", "") + "-secrets-remote-requests-writes-summary-compare.json"
	}
	if cfg.AddOnSecretsRemote.RequestsWritesSummaryCompareJSONS3Key == "" {
		cfg.AddOnSecretsRemote.RequestsWritesSummaryCompareJSONS3Key = path.Join(
			cfg.AddOnSecretsRemote.S3Dir,
			"writes-compare",
			filepath.Base(cfg.AddOnSecretsRemote.RequestsWritesSummaryCompareJSONPath),
		)
	}
	if cfg.AddOnSecretsRemote.RequestsWritesSummaryCompareTablePath == "" {
		cfg.AddOnSecretsRemote.RequestsWritesSummaryCompareTablePath = strings.ReplaceAll(cfg.ConfigPath, ".yaml", "") + "-secrets-remote-requests-writes-summary-compare.txt"
	}
	if cfg.AddOnSecretsRemote.RequestsWritesSummaryCompareTableS3Key == "" {
		cfg.AddOnSecretsRemote.RequestsWritesSummaryCompareTableS3Key = path.Join(
			cfg.AddOnSecretsRemote.S3Dir,
			"writes-compare",
			filepath.Base(cfg.AddOnSecretsRemote.RequestsWritesSummaryCompareTablePath),
		)
	}

	if cfg.AddOnSecretsRemote.RequestsReadsRawJSONPath == "" {
		cfg.AddOnSecretsRemote.RequestsReadsRawJSONPath = strings.ReplaceAll(cfg.ConfigPath, ".yaml", "") + "-secrets-remote-requests-reads-raw.json"
	}
	if cfg.AddOnSecretsRemote.RequestsReadsRawJSONS3Key == "" {
		cfg.AddOnSecretsRemote.RequestsReadsRawJSONS3Key = path.Join(
			cfg.AddOnSecretsRemote.S3Dir,
			"reads-raw",
			filepath.Base(cfg.AddOnSecretsRemote.RequestsReadsRawJSONPath),
		)
	}
	if cfg.AddOnSecretsRemote.RequestsReadsSummaryJSONPath == "" {
		cfg.AddOnSecretsRemote.RequestsReadsSummaryJSONPath = strings.ReplaceAll(cfg.ConfigPath, ".yaml", "") + "-secrets-remote-requests-reads-summary.json"
	}
	if cfg.AddOnSecretsRemote.RequestsReadsSummaryJSONS3Key == "" {
		cfg.AddOnSecretsRemote.RequestsReadsSummaryJSONS3Key = path.Join(
			cfg.AddOnSecretsRemote.S3Dir,
			"reads-summary",
			filepath.Base(cfg.AddOnSecretsRemote.RequestsReadsSummaryJSONPath),
		)
	}
	if cfg.AddOnSecretsRemote.RequestsReadsSummaryTablePath == "" {
		cfg.AddOnSecretsRemote.RequestsReadsSummaryTablePath = strings.ReplaceAll(cfg.ConfigPath, ".yaml", "") + "-secrets-remote-requests-reads-summary.txt"
	}
	if cfg.AddOnSecretsRemote.RequestsReadsSummaryTableS3Key == "" {
		cfg.AddOnSecretsRemote.RequestsReadsSummaryTableS3Key = path.Join(
			cfg.AddOnSecretsRemote.S3Dir,
			"reads-summary",
			filepath.Base(cfg.AddOnSecretsRemote.RequestsReadsSummaryTablePath),
		)
	}
	if cfg.AddOnSecretsRemote.RequestsReadsSummaryS3Dir == "" {
		cfg.AddOnSecretsRemote.RequestsReadsSummaryS3Dir = path.Join("add-on-secrets-remote", "reads-summary", cfg.Parameters.Version)
	}
	if cfg.AddOnSecretsRemote.RequestsReadsSummaryCompareJSONPath == "" {
		cfg.AddOnSecretsRemote.RequestsReadsSummaryCompareJSONPath = strings.ReplaceAll(cfg.ConfigPath, ".yaml", "") + "-secrets-remote-requests-reads-summary-compare.json"
	}
	if cfg.AddOnSecretsRemote.RequestsReadsSummaryCompareJSONS3Key == "" {
		cfg.AddOnSecretsRemote.RequestsReadsSummaryCompareJSONS3Key = path.Join(
			cfg.AddOnSecretsRemote.S3Dir,
			"reads-compare",
			filepath.Base(cfg.AddOnSecretsRemote.RequestsReadsSummaryCompareJSONPath),
		)
	}
	if cfg.AddOnSecretsRemote.RequestsReadsSummaryCompareTablePath == "" {
		cfg.AddOnSecretsRemote.RequestsReadsSummaryCompareTablePath = strings.ReplaceAll(cfg.ConfigPath, ".yaml", "") + "-secrets-remote-requests-reads-summary-compare.txt"
	}
	if cfg.AddOnSecretsRemote.RequestsReadsSummaryCompareTableS3Key == "" {
		cfg.AddOnSecretsRemote.RequestsReadsSummaryCompareTableS3Key = path.Join(
			cfg.AddOnSecretsRemote.S3Dir,
			"reads-compare",
			filepath.Base(cfg.AddOnSecretsRemote.RequestsReadsSummaryCompareTablePath),
		)
	}

	if cfg.AddOnSecretsRemote.RequestsWritesSummaryOutputNamePrefix == "" {
		cfg.AddOnSecretsRemote.RequestsWritesSummaryOutputNamePrefix = "secrets-writes-" + randutil.String(10)
	}
	if cfg.AddOnSecretsRemote.RequestsReadsSummaryOutputNamePrefix == "" {
		cfg.AddOnSecretsRemote.RequestsReadsSummaryOutputNamePrefix = "secrets-reads-" + randutil.String(10)
	}

	return nil
}
