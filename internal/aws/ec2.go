package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2Client is the interface for the EC2 client
type EC2Client interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

// EC2Service is a struct that holds the EC2 client
type EC2Service struct {
	Client EC2Client
}

// NewEC2Service creates a new EC2 service
func NewEC2Service(ctx context.Context, cfg *BaseService) (*EC2Service, error) {
	return &EC2Service{Client: ec2.NewFromConfig(cfg.Config)}, nil
}

// NewEC2ServiceWithClient creates a new EC2 service with a client
func NewEC2ServiceWithClient(client EC2Client) *EC2Service {
    return &EC2Service{Client: client}
}

// GetEC2Instances fetches EC2 instances and returns them directly
func (svc *EC2Service) GetEC2Instances(ctx context.Context) ([]types.Instance, error) {
	output, err := svc.Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}

	var instances []types.Instance
	for _, reservation := range output.Reservations {
		instances = append(instances, reservation.Instances...)
	}
	return instances, nil
}
