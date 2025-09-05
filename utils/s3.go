package utils

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

func UploadToS3(ctx context.Context, config *Config, key string, data []byte, mime string) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			config.Aws.AccessKey,
			config.Aws.SecretKey,
			"",
		),
		Region: &config.Aws.Region,
	})
	if err != nil {
		return "", errors.Wrap(err, "create aws session")
	}

	svc := s3.New(sess)

	_, err = svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(config.Aws.Bucket),
		Key:         aws.String(key),
		Body:        aws.ReadSeekCloser(bytes.NewReader(data)),
		ACL:         aws.String("public-read"),
		ContentType: aws.String(mime),
	})
	if err != nil {
		return "", errors.Wrap(err, "upload to s3")
	}

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", config.Aws.Bucket, key), nil
}
