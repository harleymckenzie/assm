#!/usr/bin/env python
"""
SSM Session Manager Helper

This script is a helper to connect to an instance using the SSM session
manager. It provides a menu to select an instance and then provides options to
connect to the instance using SSH, start an SSM session, send a command to the
instance, or print the instance ID.
"""

import subprocess
import os
import re
import signal
import logging
import contextlib
import json
from importlib.metadata import version
from datetime import datetime
from argparse import ArgumentParser
import botocore
from simple_term_menu import TerminalMenu
from boto3 import Session


def main():
    """
    Parse command line arguments and execute the appropriate actions.

    If no arguments are provided, the script will display a menu to
    select an instance and then provide options to connect to the instance
    using SSH, start an SSM session, send a command to the instance, or print
    the instance ID.
    """
    try:
        parser = arg_parser()
        args = parser.parse_args()
        logging.basicConfig(level=logging.WARNING)
        logging.getLogger("botocore").setLevel(logging.ERROR)

        if args.verbose:
            logging.basicConfig(level=logging.DEBUG)

        session = create_session(args.profile)

        if args.action == "ssm":
            start_ssm_session(session, args.instance_id)
        elif args.action == "ssh":
            ssh_to_instance(
                session, args.instance_id, args.user, args.identity
            )
        elif args.action == "cmd":
            send_command(session, args.instance_id, args.command)
        else:
            menu(session)
    except KeyboardInterrupt:
        print("Exiting...")
        exit(0)
    except botocore.exceptions.BotoCoreError as e:
        print(f"Error: {e}")
        exit(1)


def menu_create(items):
    """
    Create a menu to select an instance

    :param items: The list of items to display in the menu
    :return: The terminal_menu.chosen_accept_key, The selected item
    """
    item_details = {
        item["id"]: "\n".join(
            [
                f"ID: {item['id']}",
                f"Uptime: {item['uptime']}",
                f"Image ID: {item['image_id']}",
                f"Instance Type: {item['instance_type']}",
                f"Private IP: {item['private_ip']}",
                f"State: {item['state']}",
                f"VPC ID: {item['vpc_id']}",
                f"Subnet ID: {item['subnet_id']}",
            ]
        )
        for item in items
    }

    # Nested function to generate the preview text
    def menu_generate_preview(index):
        match = re.search(r"\((.*?)\)", index)

        if match:
            instance_id = match.group(1)
        else:
            print("No match found")

        preview = item_details.get(instance_id, "")
        shortcut_color = "\033[93m"
        quit_color = "\033[91m"
        reset_color = "\033[0m"

        shortcuts = [
            f"{shortcut_color}[M]{reset_color} Connect via SSM Session",
            f"{shortcut_color}[S]{reset_color} Connect via SSH",
            f"{shortcut_color}[P]{reset_color} Print ID",
            f"{shortcut_color}[C]{reset_color} Send Command",
            f"{quit_color}[Q]{reset_color} Quit",
        ]

        shortcut_text = "  ".join(shortcuts)
        # info_message = info_messages.get(instance_id, "")
        return f"{preview}\n\n" f"{reset_color} {shortcut_text}"

    menu_options = [f"{item['name']} ({item['id']})" for item in items]
    terminal_menu = TerminalMenu(
        menu_options,
        title="Select an instance",
        accept_keys=["m", "s", "p", "c", "enter"],
        preview_command=menu_generate_preview,  # Use the nested function
        preview_size=0.75,
        cycle_cursor=True,
        clear_screen=True,
        show_search_hint=True,
    )

    menu_entry_index = terminal_menu.show()
    if menu_entry_index is not None:
        selected_item = items[menu_entry_index]
        selection = selected_item["id"]
    else:
        print("No selection was made.")
        selection = None

    return terminal_menu.chosen_accept_key, selection


def menu(session):
    """
    Create the main menu.
    If a selection is made using one of the custom accept keys, handle the
    selection. If no selection is made, call the handle selection function.

    :param session: The AWS session
    """
    instance_ids = get_instance_ids(session)

    instance_info = get_instance_info(session, instance_ids)
    action, instance_id = menu_create(instance_info)

    if action == "m":
        start_ssm_session(session, instance_id)
    elif action == "s":
        ssh_to_instance(session, instance_id)
    elif action == "p":
        print_instance(instance_id)
    elif action == "c":
        user_command = input("Enter the command to run: ")
        send_command(session, instance_id, user_command)
    elif action == "enter":
        handle_selection(session, instance_id)


def handle_selection(session, instance_id):
    """
    Handle the selected action for a given instance.

    :param session: The AWS session
    :param instance_id: The selected instance ID
    """
    actions = [
        "[s] Connect to instance (SSH)",
        "[m] Connect to instance (SSM Session)",
        "[p] Print instance ID",
        "[c] Send Command",
    ]
    action_menu = TerminalMenu(actions, title="Select an action")
    action_index = action_menu.show()

    if action_index == 0:
        ssh_to_instance(session, instance_id)
    elif action_index == 1:
        start_ssm_session(session, instance_id)
    elif action_index == 2:
        print_instance(instance_id)
    elif action_index == 3:
        user_command = input("Enter the command to run: ")
        send_command(session, instance_id, user_command)


def create_session(profile):
    """
    Create an AWS session

    :param profile: The AWS profile to use
    :return: The AWS session
    """
    session_params = {}

    if profile:
        session_params["profile_name"] = profile
        logging.debug("Using profile: %s", profile)

    try:
        session = Session(**session_params)
        return session
    except Exception as e:
        print(f"Error: {e}")
        print("Please check your AWS credentials and try again.")
        exit(1)


def get_instance_ids(session):
    """
    Get a list of instance ids that are online

    :param session: The AWS session
    :return: A list of instance ids
    """
    ssm = session.client("ssm")

    response = ssm.describe_instance_information()

    instances = []
    for instance in response["InstanceInformationList"]:
        if instance["PingStatus"] == "Online":
            instances.append(instance["InstanceId"])

    return instances


def get_instance_info(session, instances):
    """
    Get information about the instances

    :param session: The AWS session
    :param instances: The list of instance ids
    :return: A list of instance information
    """
    ec2_client = session.client("ec2")

    response = ec2_client.describe_instances(InstanceIds=instances)

    instance_info = []
    for reservation in response["Reservations"]:
        for instance in reservation["Instances"]:
            instance_info.append(
                {
                    "id": instance["InstanceId"],
                    "uptime": relative_time(instance["LaunchTime"]),
                    "name": next(
                        (
                            tag["Value"]
                            for tag in instance["Tags"]
                            if tag["Key"] == "Name"
                        ),
                        None,
                    ),
                    "image_id": instance["ImageId"],
                    "instance_type": instance["InstanceType"],
                    "private_ip": instance["PrivateIpAddress"],
                    "state": instance["State"]["Name"],
                    "vpc_id": instance["VpcId"],
                    "subnet_id": instance["SubnetId"],
                    "tags": instance["Tags"],
                }
            )

    return instance_info


def relative_time(time):
    """
    Get the relative time from the current time

    :param time: A datetime object to convert
    :return: The relative time
    """
    diff = datetime.now().astimezone() - time

    days = diff.days
    hours, remainder = divmod(diff.seconds, 3600)
    minutes = divmod(remainder, 60)

    # Determine which two largest measurements to display
    if days > 0:
        formatted = (
            f"{days} day{'s' if days != 1 else ''}, {hours} hour{'s' if hours != 1 else ''}, "
            f"{minutes} minute{'s' if minutes != 1 else ''}"
        )
    else:
        formatted = (
            f"{hours} hour{'s' if hours != 1 else ''}, "
            f"{minutes} minute{'s' if minutes != 1 else ''}"
        )

    return formatted


@contextlib.contextmanager
def ignore_user_entered_signals():
    """
    Ignore signals sent by the user, such as Ctrl+C and Ctrl+Z.
    """
    signal_list = [signal.SIGINT, signal.SIGQUIT, signal.SIGTSTP]
    actual_signals = []
    for user_signal in signal_list:
        actual_signals.append(signal.signal(user_signal, signal.SIG_IGN))
    try:
        yield
    finally:
        for sig, user_signal in enumerate(signal_list):
            signal.signal(user_signal, actual_signals[sig])


def ssh_to_instance(
    session, instance_id, user=os.environ["USER"], identity=None
):
    """
    SSH into an instance using the provided session and instance ID.

    :param session: The AWS session
    :param instance_id: The instance ID to SSH into
    """
    # Start an SSM session and get the necessary data for ProxyCommand
    session_response, endpoint_url = create_ssm_session(
        session, instance_id, document_name="AWS-StartSSHSession"
    )

    # Construct the ProxyCommand with dynamic session data
    proxy_command = [
        "session-manager-plugin",
        json.dumps(session_response),
        session.region_name,
        "StartSession",
        session.profile_name,
        json.dumps({"Target": instance_id}),
        endpoint_url,
    ]
    # Convert the proxy command to a string, properly escaped for inclusion
    # in the SSH command
    proxy_command_str = " ".join(
        ['"' + arg.replace('"', '\\"') + '"' for arg in proxy_command]
    )

    # Specify the user and identity file if provided
    ssh_command = [
        "ssh",
        "-o",
        f"ProxyCommand={proxy_command_str}",
        *(["-i", identity] if identity else []),
        f"{user}@{instance_id}",
    ]

    logging.debug("SSH command: %s", " ".join(ssh_command))
    try:
        subprocess.check_call(ssh_command, shell=False)
    except subprocess.CalledProcessError as e:
        logging.debug("Error: %s", e)
        close_ssm_session(session, session_response["SessionId"])
        exit(1)
    close_ssm_session(session, session_response["SessionId"])


def start_ssm_session(session, instance_id):
    """
    Start an SSM session for the specified instance.

    :param session: The AWS session
    :param instance_id: The instance ID for the SSM session
    """
    session_response, endpoint_url = create_ssm_session(
        session, instance_id, document_name="SSM-SessionManagerRunShell"
    )

    command = [
        "session-manager-plugin",
        json.dumps(session_response),
        session.region_name,
        "StartSession",
        session.profile_name,
        json.dumps(dict(Target=instance_id)),
        endpoint_url,
    ]

    try:
        with ignore_user_entered_signals():
            logging.debug("Start SSM session command: %s", " ".join(command))
            subprocess.check_call(command)
        return 0
    except OSError as ex:
        if ex.errno == errno.ENOENT:
            close_ssm_session(session, session_response["SessionId"])
            raise ValueError(
                "The session-manager-plugin executable could not be found."
            ) from ex
    finally:
        close_ssm_session(session, session_response["SessionId"])


def close_ssm_session(session, session_id):
    """
    Close the SSM session

    :param session: The AWS session
    :param session_id: The session ID to close
    """
    ssm_client = session.client("ssm")
    ssm_client.terminate_session(SessionId=session_id)


def create_ssm_session(session, instance_id, document_name):
    """
    Create an SSM session for the specified instance. Optionally, create an
    SSH session.

    :param session: The AWS session
    :param instance_id: The instance ID for the SSM session
    :param ssh: Boolean indicating if the session is for SSH
    """
    ssm_client = session.client("ssm")

    try:
        response = (
            ssm_client.start_session(
                Target=instance_id, DocumentName=document_name
            )
        )
    except ssm_client.exceptions.ClientError as e:
        print(
            f"{e}\n\nPlease check your permissions and ensure the instance "
            "has appropriate permissions to use Session Manager."
        )
        exit(1)

    return (response, ssm_client.meta.endpoint_url)


def send_command(session, instance_id, user_command):
    """
    Send a command to an instance

    :param session: The AWS session
    :param instance_id: The instance id
    :param user_command: The command to run
    """
    ssm_client = session.client("ssm")
    response = ssm_client.send_command(
        InstanceIds=[instance_id],
        DocumentName="AWS-RunShellScript",
        Parameters={"commands": [user_command]},
        OutputS3BucketName="ssm-command-output",
        OutputS3KeyPrefix="output",
    )
    command_id = response["Command"]["CommandId"]

    waiter = ssm_client.get_waiter("command_executed")
    waiter.wait(
        CommandId=command_id,
        InstanceId=instance_id,
    )

    output = ssm_client.get_command_invocation(
        CommandId=command_id,
        InstanceId=instance_id,
    )
    print(output["StandardOutputContent"])


def print_instance(instance_id):
    """
    Print the specified instance ID.

    :param instance_id: The instance ID to print
    """
    print(instance_id)


def arg_parser():
    """
    Create and return the argument parser.
    """

    def file_exists(file_path):
        if not os.path.exists(file_path):
            raise FileNotFoundError(f"File not found: {file_path}")
        return file_path

    parser = ArgumentParser(
        description="Connect to an instance using the SSM session manager"
    )
    parser.add_argument("--version", action="version", version=version("assm"))
    parser.add_argument("--profile", "-p", help="The AWS profile to use")
    parser.add_argument(
        "--verbose", "-v", action="store_true", help="Enable verbose logging"
    )

    # Create a subparsers object
    subparsers = parser.add_subparsers(dest="action", help="Available actions")

    # SSM SSH Session
    ssh_parser = subparsers.add_parser(
        "ssh", help="Connect to the instance using SSH"
    )
    ssh_parser.add_argument(
        "instance_id", help="The instance ID to connect to"
    )
    ssh_parser.add_argument(
        "-u",
        "--user",
        help="The user to connect as",
        default=os.environ["USER"],
    )
    ssh_parser.add_argument(
        "-i",
        "--identity",
        help="The identity file to use for SSH",
        metavar="FILE",
        default=None,
        type=file_exists,
    )

    # SSM Session Subparser
    ssm_parser = subparsers.add_parser(
        "ssm", help="Start an SSM session on the instance"
    )
    ssm_parser.add_argument(
        "instance_id", help="The instance ID to start the session on"
    )

    # Command Subparser
    cmd_parser = subparsers.add_parser(
        "cmd", help="Send a command to the instance"
    )
    cmd_parser.add_argument(
        "instance_id", help="The instance ID to send the command to"
    )
    cmd_parser.add_argument(
        "command",
        help="The command to send to the instance",
        metavar="COMMAND",
    )

    return parser


if __name__ == "__main__":
    main()
