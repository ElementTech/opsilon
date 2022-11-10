package pull

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

func Compare(file string) *s3.GetBucketAclOutput {
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
	output, err := client.GetBucketAcl(context.TODO(), &s3.GetBucketAclInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("Owner:", *output.Owner.DisplayName)
	fmt.Println("")
	fmt.Println("Grants")

	for _, g := range output.Grants {
		// If we add a canned ACL, the name is nil
		if g.Grantee.DisplayName == nil {
			fmt.Println("  Grantee:    EVERYONE")
		} else {
			fmt.Println("  Grantee:   ", *g.Grantee.DisplayName)
		}

		fmt.Println("  Type:      ", string(g.Grantee.Type))
		fmt.Println("  Permission:", string(g.Permission))
		fmt.Println("")
	}

	fmt.Println("Your Configuration is: ", data)
	fmt.Println("Actual Configuration is: ", output)
	return output
}
