package destroy

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gopkg.in/yaml.v2"
)

func Destroy(file string) *s3.DeleteBucketOutput {
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

	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := client.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	return output
}
