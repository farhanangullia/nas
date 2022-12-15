import json
import logging
import os

import boto3
from boto3.dynamodb.types import TypeDeserializer
from boto3.session import Session
from botocore.exceptions import ClientError

logger = logging.getLogger(__name__)
logger.setLevel("INFO")

IP_SET_CONFIG_SSM_NAME = os.environ["IP_SET_CONFIG_SSM_NAME"]
ALLOW_LIST_TABLE_NAME = os.environ["ALLOW_LIST_TABLE_NAME"]

ssm_client = boto3.client("ssm")

dynamodb_resource = boto3.resource("dynamodb")
allow_list_table = dynamodb_resource.Table(ALLOW_LIST_TABLE_NAME)


class CheckRoleExistError(Exception):
    pass


def start_session(
    aws_access_key_id=None,
    aws_secret_access_key=None,
    aws_session_token=None,
    region_name=None,
):
    return boto3.Session(
        aws_access_key_id=aws_access_key_id,
        aws_secret_access_key=aws_secret_access_key,
        aws_session_token=aws_session_token,
        region_name=region_name,
    )


def assume_role(
    role_arn: str,
    role_session_name: str = "NASSession",
    session: Session = None,
):
    if not session:
        session = boto3._get_default_session()
    sts = session.client("sts")
    return sts.assume_role(
        RoleArn=role_arn, RoleSessionName=role_session_name, DurationSeconds=900
    )["Credentials"]


def check_role_exist(role_arn: str, session: Session = None):
    try:
        return assume_role(
            role_arn=role_arn,
            session=session,
        )
    except Exception as e:
        raise CheckRoleExistError("Error checking target role exist: {}".format(e))


def ddb_deserialize(r, type_deserializer=TypeDeserializer()):
    return type_deserializer.deserialize({"M": r})


def lambda_handler(event, context):
    logger.info("Started ip-expiry-handler...")
    logger.info("event: %s", str(event))

    # Deserialize dynamodb items
    ips = [ddb_deserialize(r["dynamodb"]["OldImage"]) for r in event["Records"]]
    for ip in ips:
        ip_address = ip["IpAddress"]
        aws_account_id = ip["AwsAccountId"]

        logger.info("ip_address: %s", str(ip_address))
        logger.info("aws_account_id: %s", str(aws_account_id))

        #### Retrieve WAF IpSet config from SSM Parameter Store ####
        logger.info(
            "PARAM: %s",
            str(
                ssm_client.get_parameter(Name=IP_SET_CONFIG_SSM_NAME)["Parameter"][
                    "Value"
                ]
            ),
        )
        waf_ipset_config = json.loads(
            ssm_client.get_parameter(Name=IP_SET_CONFIG_SSM_NAME)["Parameter"]["Value"]
        )
        logger.info("waf_ipset_config: %s", str(waf_ipset_config))

        #### Check if IAM role in target account exist ####
        logger.info("Checking role exist...")
        credentials = check_role_exist(waf_ipset_config[aws_account_id]["IamRole"])
        logger.info("Role exists...")

        #### Create WAFv2 session in target account ####
        target_account_session = start_session(
            aws_access_key_id=credentials["AccessKeyId"],
            aws_secret_access_key=credentials["SecretAccessKey"],
            aws_session_token=credentials["SessionToken"],
        )
        wafv2_client = target_account_session.client(
            "wafv2", region_name="us-east-1"
        )  # CloudFront IPSet needs to be in us-east-1

        #### Get IPSet ####
        logger.info(
            "waf_ipset_config for target account: %s",
            str(waf_ipset_config[aws_account_id]),
        )
        get_response = wafv2_client.get_ip_set(
            Name=waf_ipset_config[aws_account_id]["Name"],
            Scope="CLOUDFRONT",
            Id=waf_ipset_config[aws_account_id]["Id"],
        )

        #### Remove expired IP ####
        addresses = get_response["IPSet"]["Addresses"]
        if ip_address in addresses:
            addresses.remove(ip_address)

        try:
            logger.info(
                "Updating WAF IpSet for: %s with %s",
                str(waf_ipset_config[aws_account_id]),
                str(addresses),
            )

            update_response = wafv2_client.update_ip_set(
                Name=waf_ipset_config[aws_account_id]["Name"],
                Scope="CLOUDFRONT",
                Id=waf_ipset_config[aws_account_id]["Id"],
                Description=get_response["IPSet"]["Description"],
                Addresses=addresses,
                LockToken=get_response["LockToken"],
            )
            # If success, update request to completed
            logger.info("Successfully updated WAF IPSet")
        except Exception as e:
            logger.error(
                "Error occurred when updating waf ipset: %s", str(e), exc_info=1
            )
            continue

    logger.info("Ended ip-expiry-handler...")
    return
