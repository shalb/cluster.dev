package aws

import (
	"bytes"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func S3Put(region, bucketName, keyName, data string) error {
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(region),
		},
	)
	if err != nil {
		return err
	}
	svc := s3.New(sess)
	input := &s3.PutObjectInput{
		Body:   strings.NewReader(data),
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	}
	// log.Debugf("S3 uploading object. Bucket '%v', key: '%v', region: '%v'", bucketName, keyName, region)
	// Create an uploader with the session and default options
	_, err = svc.PutObject(input)
	if err != nil {
		return err
	}
	return err
}

func S3Get(region, bucketName, keyName string) (string, error) {
	svc := s3.New(session.New(
		&aws.Config{
			Region: aws.String(region),
		},
	))
	input := &s3.GetObjectInput{

		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	}
	// log.Debugf("S3 downloading object. Bucket '%v', key: '%v', region: '%v'", bucketName, keyName, region)
	// Create an uploader with the session and default options
	result, err := svc.GetObject(input)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(result.Body)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func S3Delete(region, bucketName, keyName string) error {
	svc := s3.New(session.New(
		&aws.Config{
			Region: aws.String(region),
		},
	))
	// log.Debugf("S3 deleting object. Bucket '%v', key: '%v', region: '%v'", bucketName, keyName, region)
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	})

	if err != nil {
		return err
	}
	return nil
}
