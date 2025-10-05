package aws

import (
    "context"
    "testing"

	"github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Mock EC2 client
type mockEC2Client struct {
    DescribeInstancesFunc func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
    return m.DescribeInstancesFunc(ctx, params, optFns...)
}

func TestGetEC2Instances(t *testing.T) {
    mockClient := &mockEC2Client{
		DescribeInstancesFunc: func(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
            return &ec2.DescribeInstancesOutput{
                Reservations: []types.Reservation{
                    {
                        Instances: []types.Instance{
                            {InstanceId: aws.String("i-123")},
                            {InstanceId: aws.String("i-456")},
                        },
                    },
                },
            }, nil
		},
	}

	svc := NewEC2ServiceWithClient(mockClient)
	instances, err := svc.GetEC2Instances(context.Background())

	if err != nil {
		t.Fatalf("expected 2 instances, got %v", err)
	}
	if len(instances) != 2 {
        t.Errorf("expected 2 instances, got %d", len(instances))
    }
}
