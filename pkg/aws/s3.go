package aws

import (
	"strings"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func S3Upload(region, bucketName, keyName, data string) error {
	// Upload input parameters
	r := strings.NewReader(data)
	upParams := &s3manager.UploadInput{
		Bucket: &bucketName,
		Key:    &keyName,
		Body:   r,
	}
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession(
		&aws.Config{
			Region: aws.String(region),
		},
	))
	log.Debugf("S3 state lock. Bucket '%v', key: '%v', region: '%v'", bucketName, keyName, region)
	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)
	// Perform an upload.
	result, err := uploader.Upload(upParams)
	if err == nil {
		if result.VersionID != nil {
			log.Debugf("S3 upload result: %v", result.VersionID)
		}
	} else {
		log.Debug(err.Error())
	}
	// Perform upload with options different than the those in the Uploader.

	return err
}
