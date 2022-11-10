package push

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"gopkg.in/yaml.v2"
)

func Create(file string) *s3.CreateBucketOutput {
	yfile, err := ioutil.ReadFile(file)

	if err != nil {

		log.Fatal(err)
	}

	data := make(map[string]interface{})

	err2 := yaml.Unmarshal(yfile, &data)

	if err2 != nil {

		log.Fatal(err2)
	}

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	var name string = fmt.Sprint(data["name"])
	// var objectLockEnabledForBucket bool = data["objectLockEnabledForBucket"].(bool)

	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket:                    aws.String(name),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{LocationConstraint: types.BucketLocationConstraintEuWest1},
	})
	client.PutObjectLockConfiguration(context.TODO(), &s3.PutObjectLockConfigurationInput{
		Bucket:              aws.String(name),
		ChecksumAlgorithm:   "",
		ContentMD5:          new(string),
		ExpectedBucketOwner: new(string),
		ObjectLockConfiguration: &types.ObjectLockConfiguration{
			ObjectLockEnabled: "",
			Rule:              &types.ObjectLockRule{},
		},
		RequestPayer: "",
		Token:        new(string),
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	return output
}
