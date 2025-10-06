package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/harleymckenzie/assm/internal/apperror"
	"github.com/harleymckenzie/assm/internal/aws"
	"github.com/harleymckenzie/assm/internal/table"
	"github.com/spf13/cobra"
)

// Global configuration variables
var (
	Version = "0.0.3"
)

var (
	rootCmd = &cobra.Command{
		Use:     "assm [instance id]",
		Short:   "A tool to manage and connect to EC2 instances",
		Version: Version,
		// Allow an optional "instance id" arg to connect directly to the instance
		// CompletionOptions: cobra.CompletionOptions{
		// 	DisableDefaultCmd: true,
		// },
		Run: func(cmd *cobra.Command, args []string) {
			// 1. Check and confirm the session manager plugin has been installed
			if err := verifyPlugin(); err != nil {
				apperror.Exit(apperror.New(apperror.CodePluginNotFound, fmt.Errorf("verify session-manager-plugin: %w", err)))
			}

			ctx := context.TODO()
			profile, region := getPersistentFlags(cmd)
			awsCfg, err := aws.LoadDefaultConfig(ctx, profile, region)
			if err != nil {
				apperror.Exit(apperror.New(apperror.CodeGeneralError, fmt.Errorf("load config: %w", err)))
			}

			// 2. Create the service clients
			ec2Client, err := aws.NewEC2Service(ctx, awsCfg)
			if err != nil {
				apperror.Exit(apperror.New(apperror.CodeGeneralError, fmt.Errorf("create ec2 service: %w", err)))
			}

			ssmClient, err := aws.NewSSMService(ctx, awsCfg)
			if err != nil {
				apperror.Exit(apperror.New(apperror.CodeGeneralError, fmt.Errorf("create ssm service: %w", err)))
			}

			// 3. Get EC2 instances
			instances, err := ec2Client.GetEC2Instances(ctx)
			if err != nil {
				apperror.Exit(apperror.New(apperror.CodeGeneralError, fmt.Errorf("get ec2 instances: %w", err)))
			}

			// 4. Format into Rows
			rows, err := table.BuildRows(instances)
			if err != nil {
				apperror.Exit(apperror.New(apperror.CodeGeneralError, fmt.Errorf("build rows: %w", err)))
			}
			// 5. Print table output and return selected instance id
			instanceId := table.ShowTableAndSelect(rows)
			if instanceId == "" {
				fmt.Println("No instance selected. Exiting.")
				return
			}

			// 6. Create session manager session and return the session response
			sessionId, err := ssmClient.StartSession(ctx, profile, awsCfg.Config.Region, instanceId)
			if err != nil {
				apperror.Exit(apperror.New(apperror.CodeGeneralError, fmt.Errorf("start session: %w", err)))
			}

			// 7. Terminate session manager session
			err = ssmClient.TerminateSession(ctx, profile, awsCfg.Config.Region, sessionId)
			if err != nil {
				apperror.Exit(apperror.New(apperror.CodeGeneralError, fmt.Errorf("terminate session: %w", err)))
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
	rootCmd.PersistentFlags().StringP("profile", "p", "", "AWS profile to use for authentication.")
	rootCmd.PersistentFlags().StringP("region", "r", "", "AWS region.")
	rootCmd.AddCommand(completionCmd())
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
