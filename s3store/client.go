package s3store

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// newS3Client builds an S3 client from the supplied config.
//
// When AWS_ENDPOINT_URL_S3 or AWS_ENDPOINT_URL is set (typical for MinIO or
// LocalStack development), path-style addressing is enabled. Virtual-host style
// would otherwise route requests to <bucket>.<endpoint-host>, which fails to
// resolve against a local endpoint that lacks wildcard DNS.
func newS3Client(cfg aws.Config) *s3.Client {
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		if os.Getenv("AWS_ENDPOINT_URL_S3") != "" || os.Getenv("AWS_ENDPOINT_URL") != "" {
			o.UsePathStyle = true
		}
	})
}
