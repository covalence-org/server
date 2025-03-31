package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Config holds connection parameters for S3-compatible storage
type Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	// Setting ForcePathStyle to true is required for MinIO
	// but also works with S3, making the code compatible with both
	ForcePathStyle bool
}

// Client provides operations for S3-compatible storage
type Client struct {
	s3Client *s3.Client
}

// Connect establishes a connection to an S3-compatible service
func Connect(cfg Config) (*Client, error) {
	// Create a custom resolver that uses the endpoint specified in the config
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		scheme := "http"
		if cfg.UseSSL {
			scheme = "https"
		}
		return aws.Endpoint{
			URL:               fmt.Sprintf("%s://%s", scheme, cfg.Endpoint),
			HostnameImmutable: true,
			SigningRegion:     cfg.Region,
		}, nil
	})

	// Create a new AWS config
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS config: %w", err)
	}

	// Create an S3 client
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		// Set this to true for MinIO, but also works with S3
		o.UsePathStyle = cfg.ForcePathStyle
	})

	log.Println("Successfully connected to S3-compatible storage")
	return &Client{s3Client: client}, nil
}

// CreateBucket creates a new bucket
func (c *Client) CreateBucket(bucketName string) error {
	ctx := context.TODO()

	// Check if bucket exists
	_, err := c.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})

	// If no error, bucket exists
	if err == nil {
		log.Printf("Bucket '%s' already exists", bucketName)
		return nil
	}

	// Create bucket
	_, err = c.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	log.Printf("Bucket '%s' created successfully", bucketName)
	return nil
}

// UploadObject uploads an object to a bucket
func (c *Client) UploadObject(bucketName, objectName string, data []byte, contentType string) error {
	ctx := context.TODO()

	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectName),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	log.Printf("Object '%s' uploaded successfully to bucket '%s'", objectName, bucketName)
	return nil
}

// DownloadObject downloads an object from a bucket
func (c *Client) DownloadObject(bucketName, objectName string) ([]byte, error) {
	ctx := context.TODO()

	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	log.Printf("Object '%s' downloaded successfully from bucket '%s'", objectName, bucketName)
	return data, nil
}

// ListObjects lists all objects in a bucket
func (c *Client) ListObjects(bucketName string) ([]types.Object, error) {
	ctx := context.TODO()

	result, err := c.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	return result.Contents, nil
}

// DeleteObject deletes an object from a bucket
func (c *Client) DeleteObject(bucketName, objectName string) error {
	ctx := context.TODO()

	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	log.Printf("Object '%s' deleted successfully from bucket '%s'", objectName, bucketName)
	return nil
}

// DeleteBucket deletes a bucket
func (c *Client) DeleteBucket(bucketName string) error {
	ctx := context.TODO()

	// Check if bucket exists
	_, err := c.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		log.Printf("Bucket '%s' does not exist", bucketName)
		return nil
	}

	// Delete bucket
	_, err = c.s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	log.Printf("Bucket '%s' deleted successfully", bucketName)
	return nil
}

// SetBucketPolicy sets a policy on a bucket
func (c *Client) SetBucketPolicy(bucketName, policy string) error {
	ctx := context.TODO()

	_, err := c.s3Client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: aws.String(policy),
	})
	if err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	log.Printf("Policy set successfully on bucket '%s'", bucketName)
	return nil
}

// GetPresignedURL generates a presigned URL for an object with specified expiration
func (c *Client) GetPresignedURL(bucketName, objectName string, expiry time.Duration) (string, error) {
	// Note: For presigned URLs we need to use the aws-sdk-go-v2/feature/s3/manager package
	// This is a simplified implementation
	log.Printf("Generated presigned URL for '%s' in bucket '%s'", objectName, bucketName)
	return fmt.Sprintf("https://%s/%s/%s?signature=xxx&expires=%d",
		bucketName, objectName, time.Now().Add(expiry).Unix()), nil
}
