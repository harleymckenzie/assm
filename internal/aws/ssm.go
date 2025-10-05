package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type SSMClient interface {
	DescribeInstanceInformation(ctx context.Context, in *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error)
	StartSession(ctx context.Context, in *ssm.StartSessionInput, optFns ...func(*ssm.Options)) (*ssm.StartSessionOutput, error)
}

// SSMService is a struct that holds the SSM client.
type SSMService struct {
	Client  SSMClient
}

// NewSSMService creates a new SSM service.
func NewSSMService(ctx context.Context, cfg *BaseService) (*SSMService, error) {
	return &SSMService{
		Client: ssm.NewFromConfig(cfg.Config),
	}, nil
}

func NewSSMServiceWithClient(client SSMClient, region string, profile string) *SSMService {
	return &SSMService{Client: client}
}

func (svc *SSMService) StartSession(ctx context.Context, profile string, region string, instanceId string) (error) {
	fmt.Printf("Starting SSM session using target instance: %s\n", instanceId)
	out, err := svc.Client.StartSession(ctx, &ssm.StartSessionInput{
		Target:       aws.String(instanceId),
		DocumentName: aws.String("SSM-SessionManagerRunShell"),
	})
	if err != nil {
		return err
	}

	// 2) Build params JSON (what CLI passes as --parameters)
	endpointUrl := aws.ToString(out.StreamUrl)
	params := map[string]string{
		"Target": instanceId,
	}
	paramsJSON, _ := json.Marshal(params)
	fmt.Printf("Params json: %s\n", paramsJSON)

	// 3) Invoke the plugin (stdin/out wired to your TUI)
	respJSON, _ := json.Marshal(out)
	fmt.Printf("Response json: %s\n", respJSON)

	fmt.Printf("Profile: %s, Region: %s\n", profile, region)
	cmd := exec.CommandContext(ctx, "session-manager-plugin",
		string(respJSON),
		region,
		"StartSession",
		profile,
		string(paramsJSON),
		endpointUrl, // endpoint override; usually empty
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()

	return nil
}
