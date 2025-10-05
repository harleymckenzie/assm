package aws

import (
	"fmt"
	"os"
)

// ValidateAWSConfig checks that region and profile are set via CLI args or env vars
func ValidateAWSConfig(profileArg, regionArg string) (string, error) {
	// Get profile from arg or env var
	profile := profileArg
	if profile == "" {
		profile = os.Getenv("AWS_PROFILE")
	}
	
	// Validate both are set
	if profile == "" {
		return "", fmt.Errorf("AWS profile must be set via --profile flag or AWS_PROFILE environment variable")
	}
	
	return profile, nil
}