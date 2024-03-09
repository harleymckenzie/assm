# README.md

## AWS Simple Session Manager (assm)

The ASSM helper script facilitates connecting to AWS instances using the SSM Session Manager. It provides a user-friendly interface to select an instance and connect via a Session Manager SSH session, a regular SSM session, send commands, or print the instance ID.

### Features

- Connect to an instance using SSH or SSM Session.
- Send commands to an instance.
- Print the instance ID.
- User-friendly menu for instance selection.

### Installation

#### Using Homebrew

To install ASSM using Homebrew, run the following command in your terminal:

```bash
brew tap harleymckenzie/asc
brew install assm
```

#### Using pip

To install ASSM using pip, run the following command in your terminal:

```bash
pip install assm
```

### Usage
To use assm, follow these steps:

1. Ensure you have AWS CLI and SSM Session Manager Plugin installed.
2. Configure your AWS credentials using `aws configure`.
3. Run the script using only the `assm` command and select an instance from the menu, or use one of the available options.

#### Options

- `--version`: Show the version of the ASSM helper script.
- `--profile, -p`: Specify the AWS profile to use.
- `--instance-id, -id`: Specify the instance ID for the action.
- `--verbose, -v`: Enable verbose logging.

#### Actions

- `ssh`: Connect to the instance using SSH.
  - `-u, --user`: Specify the user to connect as. Defaults to the current system user.
  - `-i, --identity`: Specify the identity file to use for SSH.

- `ssm`: Start an SSM session on the instance.

- `cmd`: Send a command to the instance.

- `output`: Print the instance ID (useful for passing the instance ID from menu selection to another command).


### License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/harleymckenzie/assm/blob/main/LICENSE) file for details.