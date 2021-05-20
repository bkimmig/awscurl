package lib

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
)

// GetAWSConfig builds the AWS Config based on the provided AWS-related flags
func GetAwsConfig(profile string, accessKey string, secretKey string, sessionToken string, region string) (aws.Config, error) {
	var cfg aws.Config
	var cfgSources external.Configs

	if profile != "" {
		awsProfileLoader := external.WithSharedConfigProfile(profile)
		cfgSources = append(cfgSources, awsProfileLoader)
	}
	if accessKey != "" && secretKey != "" {
		staticCredsLoader := external.WithCredentialsProvider{
			CredentialsProvider: aws.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID: accessKey, SecretAccessKey: secretKey, SessionToken: sessionToken,
				},
			},
		}
		cfgSources = append(cfgSources, staticCredsLoader)
	}

	cfg, err := external.LoadDefaultAWSConfig(cfgSources...)
	if err != nil {
		return cfg, fmt.Errorf("Unable to load AWS config: %s", err)
	}

	if region != "" {
		cfg.Region = region
	}

	return cfg, nil
}
