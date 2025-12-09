package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// SSMClient is the interface for the SSM client
type SSMClient interface {
	DescribeInstanceInformation(ctx context.Context, in *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error)
	StartSession(ctx context.Context, in *ssm.StartSessionInput, optFns ...func(*ssm.Options)) (*ssm.StartSessionOutput, error)
	TerminateSession(ctx context.Context, in *ssm.TerminateSessionInput, optFns ...func(*ssm.Options)) (*ssm.TerminateSessionOutput, error)
}

// SSMService is a struct that holds the SSM client
type SSMService struct {
	Client SSMClient
}

// NewSSMService creates a new SSM service
func NewSSMService(ctx context.Context, cfg *BaseService) (*SSMService, error) {
	return &SSMService{
		Client: ssm.NewFromConfig(cfg.Config),
	}, nil
}

// NewSSMServiceWithClient creates a new SSM service with a client
func NewSSMServiceWithClient(client SSMClient, region string, profile string) *SSMService {
	return &SSMService{Client: client}
}

// StartSession starts a new SSM session and returns the session ID
func (svc *SSMService) StartSession(ctx context.Context, profile string, region string, instanceId string) (string, error) {
	out, err := svc.Client.StartSession(ctx, &ssm.StartSessionInput{
		Target:       aws.String(instanceId),
		DocumentName: aws.String("SSM-SessionManagerRunShell"),
	})
	if err != nil {
		return "", err
	}

	// 2) Build params JSON (what CLI passes as --parameters)
	endpointUrl := aws.ToString(out.StreamUrl)
	sessionId := aws.ToString(out.SessionId)
	params := map[string]string{
		"Target": instanceId,
	}
	paramsJSON, _ := json.Marshal(params)

	// 3) Invoke the plugin
	respJSON, _ := json.Marshal(out)

	fmt.Printf("Connecting to instance %s using session ID: %s\n", instanceId, sessionId)
	cmd := exec.Command("session-manager-plugin",
		string(respJSON),
		region,
		"StartSession",
		profile,
		string(paramsJSON),
		endpointUrl,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set up signal handling to forward SIGINT to the plugin process
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start plugin: %w", err)
	}

	// Handle signals by forwarding them to the plugin process
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Forward any signals to the plugin process and wait for it to finish
	for {
		select {
		case sig := <-sigChan:
			if cmd.Process != nil {
				// Forward the signal to the plugin process instead of terminating our process
				_ = cmd.Process.Signal(sig)
			}
		case err = <-done:
			// Command finished (either normally or after signal)
			return sessionId, err
		}
	}
}

func (svc *SSMService) TerminateSession(ctx context.Context, profile string, region string, sessionId string) error {
	_, err := svc.Client.TerminateSession(ctx, &ssm.TerminateSessionInput{
		SessionId: aws.String(sessionId),
	})
	if err != nil {
		return fmt.Errorf("terminate session: %w", err)
	}
	fmt.Printf("Terminated SSM session %s\n", sessionId)
	return nil
}
