package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/harleymckenzie/assm/internal/aws"
	"github.com/harleymckenzie/assm/internal/exitcode"
	"github.com/harleymckenzie/assm/internal/table"
	"github.com/spf13/cobra"
)

// Global configuration variables
var (
	Version = "0.0.1"
)

var (
	rootCmd = &cobra.Command{
		Use:     "assm",
		Short:   "A tool to manage and connect to EC2 instances",
		Version: Version,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			// 1. Check and confirm the session manager plugin has been installed
			if err := verifyPlugin(); err != nil {
				exitcode.New(exitcode.CodeUnknownCommand, fmt.Errorf("verify session-manager-plugin: %w", err))
			}

			ctx := context.TODO()
			profile, region := getPersistentFlags(cmd)
			awsCfg, err := aws.LoadDefaultConfig(ctx, profile, region)
			if err != nil {
				exitcode.New(exitcode.CodeGeneralError, fmt.Errorf("load AWS config: %w", err))
			}
			
			// 2. Create the service clients
			ec2Client, err := aws.NewEC2Service(ctx, awsCfg)
			if err != nil {
				exitcode.New(exitcode.CodeGeneralError, fmt.Errorf("create ec2 service: %w", err))
			}
			
			ssmClient, err := aws.NewSSMService(ctx, awsCfg)
			if err != nil {
				exitcode.New(exitcode.CodeGeneralError, fmt.Errorf("create ssm service: %w", err))
			}

			// 3. Get EC2 instances
			instances, err := ec2Client.GetEC2Instances(ctx)
			if err != nil {
				exitcode.New(exitcode.CodeGeneralError, fmt.Errorf("get ec2 instances: %w", err))
			}

			// 4. Format into Rows
			rows, err := table.BuildRows(instances)
			if err != nil {
				exitcode.New(exitcode.CodeGeneralError, fmt.Errorf("build rows: %w", err))
			}
			// 5. Print table output and return selected instance id
			instanceId := table.Render(rows)
			fmt.Printf("Returned instance id: %s\n", instanceId)

			// 6. Create session manager session and return the session response
			err = ssmClient.StartSession(ctx, profile, awsCfg.Config.Region, instanceId)
			if err != nil {
				exitcode.New(exitcode.CodeGeneralError, fmt.Errorf("start session: %w", err))
			}
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("profile", "p", "", "AWS profile to use for authentication. Will be used as client-id if no client-id is provided.")
	rootCmd.PersistentFlags().StringP("region", "r", "", "AWS region.")
}

func getPersistentFlags(cmd *cobra.Command) (string, string) {
	profile, _ := cmd.Root().PersistentFlags().GetString("profile")
	region, _ := cmd.Root().PersistentFlags().GetString("region")
	return profile, region
}

// verifyPlugin ensures session-manager-plugin is installed and callable
func verifyPlugin() error {
	cmd := exec.Command("session-manager-plugin")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New("session-manager-plugin not found; install it from AWS CLI docs")
	}
	if !strings.Contains(string(out), "The Session Manager plugin was installed successfully") {
		return errors.New("session-manager-plugin output unexpected: " + string(out))
	}
	return nil
}
