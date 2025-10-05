package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type BaseService struct {
	Config aws.Config
}

func LoadDefaultConfig(ctx context.Context, profile string, region string) (*BaseService, error) {
	opts := []func(*config.LoadOptions) error{}

	if profile == "" {
		profile = os.Getenv("AWS_PROFILE")
	}

	if profile == "" {
		return nil, fmt.Errorf("AWS profile must be set via --profile flag or AWS_PROFILE environment variable")
	}

	opts = append(opts, config.WithSharedConfigProfile(profile))

	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &BaseService{Config: cfg}, nil
}
