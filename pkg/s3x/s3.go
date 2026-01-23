package s3x

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
)

func NewClient(ctx context.Context, options *Options) (*s3.Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				options.GetAccessKeyId().GetValue(),
				options.GetSecretAccessKey().GetValue(),
				"",
			)),
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.AppID = options.GetAppId().GetValue()
		o.BaseEndpoint = aws.String(options.GetBaseEndpoint().GetValue())
		o.UsePathStyle = options.GetUsePathStyle().GetValue()
		o.RetryMaxAttempts = int(options.GetRetryMaxAttempts().GetValue())
	})
	_, err = s3Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(options.GetBucket().GetValue())})
	if err == nil {
		return s3Client, nil
	}

	// 如果是NoSuchBucket错误，我们可能需要创建bucket
	// 由于AWS SDK v2不直接暴露特定的错误类型，我们需要检查错误代码
	if isNoSuchBucketError(err) {
		// 这里可以添加创建bucket的逻辑
		_, createErr := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(options.GetBucket().GetValue()),
		})
		if createErr != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", createErr)
		}
	}
	return nil, fmt.Errorf("failed to head bucket: %w", err)
}

// isNoSuchBucketError 检查错误是否为NoSuchBucket错误
func isNoSuchBucketError(err error) bool {
	var apiErr interface{ ErrorCode() string }
	if errors.As(err, &apiErr) {
		return apiErr.ErrorCode() == "NoSuchBucket" || apiErr.ErrorCode() == "NotFound"
	}
	return false
}

func NewUploader(s3Client *s3.Client) *manager.Uploader {
	return manager.NewUploader(s3Client)
}

func NewDownloader(s3Client *s3.Client) *manager.Downloader {
	return manager.NewDownloader(s3Client)
}
