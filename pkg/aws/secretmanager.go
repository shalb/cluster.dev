package aws

// Use this code snippet in your app.
// If you need more information about configurations or implementing the sample code, visit the AWS docs:
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func GetSecret(region string, secretName string) (interface{}, error) {
	log.Debugf("Downloading secret %s (%s)", secretName, region)
	svc := secretsmanager.New(session.New(),
		aws.NewConfig().WithRegion(region))
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return "", aerr
		}
		return "", err
	}
	var secretData string
	if result.SecretString != nil {
		secretData = *result.SecretString

	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		len, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			fmt.Println("Base64 Decode Error:", err)
			return "", err
		}
		secretData = string(decodedBinarySecretBytes[:len])
	}
	parseCheck := []interface{}{}
	errSliceCheck := json.Unmarshal([]byte(secretData), &parseCheck)

	parsed := map[string]interface{}{}
	err = json.Unmarshal([]byte(secretData), &parsed)
	if err != nil {
		if errSliceCheck != nil {
			log.Debugf("Secret '%v' is not JSON, creating raw data", secretName)
			return secretData, nil
		}
		return nil, fmt.Errorf("aws get secret: JSON secret must be a map, not array")
	}
	return parsed, nil
}

func CreateSecret(region string, secretName string, secretData interface{}) (err error) {

	kind := reflect.TypeOf(secretData).Kind()
	var secretDataStr string

	if kind == reflect.Map {
		secretDataByte, err := json.Marshal(secretData)
		if err != nil {
			return err
		}
		secretDataStr = string(secretDataByte)
	} else {
		secretDataStr = fmt.Sprintf("%v", secretData)
	}

	if kind == reflect.Slice {
		return fmt.Errorf("create secret: array is not allowed")
	}

	svc := secretsmanager.New(session.New(),
		aws.NewConfig().WithRegion(region))
	input := &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(secretDataStr),
	}

	result, err := svc.CreateSecret(input)
	if err != nil {
		return
	}

	fmt.Println(result)
	return
}

func UpdateSecret(region string, secretName string, secretData interface{}) (err error) {

	kind := reflect.TypeOf(secretData).Kind()
	var secretDataStr string

	if kind == reflect.Map {
		secretDataByte, err := json.Marshal(secretData)
		if err != nil {
			return err
		}
		secretDataStr = string(secretDataByte)
	} else {
		secretDataStr = fmt.Sprintf("%v", secretData)
	}

	if kind == reflect.Slice {
		return fmt.Errorf("create secret: array is not allowed")
	}

	svc := secretsmanager.New(session.New(),
		aws.NewConfig().WithRegion(region))
	input := &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(secretDataStr),
	}

	result, err := svc.CreateSecret(input)
	if err != nil {
		return
	}

	fmt.Println(result)
	return
}
