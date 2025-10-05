package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type BaseService struct {
	Config aws.Config
}

func LoadDefaultConfig(ctx context.Context, profile string, region string) (*BaseService, error) {
    opts := []func(*config.LoadOptions) error{}
    
    if profile != "" {
        opts = append(opts, config.WithSharedConfigProfile(profile))
    }
    
    if region != "" {
        opts = append(opts, config.WithRegion(region))
    }
    
    cfg, err := config.LoadDefaultConfig(ctx, opts...)
    if err != nil {
		return nil, err
	}

	return &BaseService{Config: cfg}, nil
}