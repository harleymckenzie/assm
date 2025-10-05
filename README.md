# AWS Simple Session Manager (assm)

The assm CLI is designed to provide a more simplified way to connect to EC2 instances using via AWS Systems Manager Sessions Manager.

### Installation

#### Using Homebrew

To install ASSM using Homebrew, run the following command in your terminal:

```bash
brew tap harleymckenzie/asc
brew install assm
```

### Usage
To use assm, follow these steps:

1. Ensure you have AWS CLI and SSM Session Manager Plugin installed.
2. Configure your AWS credentials using `aws configure`.
3. Run the script using only the `assm` command and select an instance from the menu, or use one of the available options.

#### Options

- `--profile, -p`: Specify the AWS profile to use.
- `--instance-id, -id`: Specify the instance ID to connect to.
- `--verbose, -v`: Enable verbose logging.
- `--version`: Show the version of the ASSM helper script.

#### Actions

- `ssh`: Connect to the instance using SSH.
  - `-u, --user`: Specify the user to connect as. Defaults to the current system user.
  - `-i, --identity`: Specify the identity file to use for SSH.

- `ssm`: Start an SSM session on the instance.

- `cmd`: Send a command to the instance.

- `output`: Print the instance ID (useful for passing the instance ID from menu selection to another command).


### License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/harleymckenzie/assm/blob/main/LICENSE) file for details.# ssm-document-runner-client
