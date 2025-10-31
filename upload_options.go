package awstools

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// UploadOption allows customizing the S3 PutObjectInput before the upload.
type UploadOption func(input *s3.PutObjectInput)

// WithUploadContentType sets the Content-Type header for the uploaded object.
func WithUploadContentType(contentType string) UploadOption {
	return func(input *s3.PutObjectInput) {
		input.ContentType = aws.String(contentType)
	}
}

// WithUploadContentDisposition sets the Content-Disposition header.
func WithUploadContentDisposition(contentDisposition string) UploadOption {
	return func(input *s3.PutObjectInput) {
		input.ContentDisposition = aws.String(contentDisposition)
	}
}

// WithUploadCacheControl sets the Cache-Control header.
func WithUploadCacheControl(cacheControl string) UploadOption {
	return func(input *s3.PutObjectInput) {
		input.CacheControl = aws.String(cacheControl)
	}
}

// WithUploadContentEncoding sets the Content-Encoding header.
func WithUploadContentEncoding(contentEncoding string) UploadOption {
	return func(input *s3.PutObjectInput) {
		input.ContentEncoding = aws.String(contentEncoding)
	}
}

// WithUploadContentLanguage sets the Content-Language header.
func WithUploadContentLanguage(contentLanguage string) UploadOption {
	return func(input *s3.PutObjectInput) {
		input.ContentLanguage = aws.String(contentLanguage)
	}
}

// WithUploadMetadata sets custom metadata key/value pairs for the uploaded object.
func WithUploadMetadata(metadata map[string]string) UploadOption {
	return func(input *s3.PutObjectInput) {
		if metadata == nil {
			return
		}
		if input.Metadata == nil {
			input.Metadata = make(map[string]string, len(metadata))
		}
		for k, v := range metadata {
			input.Metadata[k] = v
		}
	}
}

// WithUploadACL applies a canned ACL to the uploaded object.
func WithUploadACL(acl types.ObjectCannedACL) UploadOption {
	return func(input *s3.PutObjectInput) {
		input.ACL = acl
	}
}
