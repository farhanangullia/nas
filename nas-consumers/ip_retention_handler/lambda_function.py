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
REQUESTS_TABLE_NAME = os.environ["REQUESTS_TABLE_NAME"]
ALLOW_LIST_TABLE_NAME = os.environ["ALLOW_LIST_TABLE_NAME"]

ssm_client = boto3.client("ssm")

dynamodb_resource = boto3.resource("dynamodb")
requests_table = dynamodb_resource.Table(REQUESTS_TABLE_NAME)
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
    logger.info("Started ip-retention-handler...")
    logger.info("event: %s", str(event))

    # Deserialize dynamodb items
    requests = [ddb_deserialize(r["dynamodb"]["NewImage"]) for r in event["Records"]]
    for request in requests:
        request_id = request["Id"]
        ip_address_obj = request["IpAddress"]
        aws_account_id = request["IpAddress"]["AwsAccountId"]
        status = request["Status"]

        logger.info("request_id: %s", str(request_id))
        logger.info("ip_address_obj: %s", str(ip_address_obj))
        logger.info("aws_account_id: %s", str(aws_account_id))
        logger.info("status: %s", str(status))

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

        # Combine existing addresses with request address
        addresses = get_response["IPSet"]["Addresses"]
        addresses.append(ip_address_obj["Ip"])
        new_addresses = set(addresses)
        new_addresses = list(new_addresses)

        try:
            logger.info(
                "Updating WAF IpSet for: %s with %s",
                str(waf_ipset_config[aws_account_id]),
                str(new_addresses),
            )

            update_response = wafv2_client.update_ip_set(
                Name=waf_ipset_config[aws_account_id]["Name"],
                Scope="CLOUDFRONT",
                Id=waf_ipset_config[aws_account_id]["Id"],
                Description=get_response["IPSet"]["Description"],
                Addresses=new_addresses,
                LockToken=get_response["LockToken"],
            )
            # If success, update request to completed
            logger.info("Successfully updated WAF IPSet")
        except Exception as e:
            logger.error(
                "Error occurred when updating waf ipset: %s", str(e), exc_info=1
            )
            try:
                ddb_response = requests_table.update_item(
                    Key={"Id": request_id},
                    UpdateExpression="set #request_status=:s",
                    ExpressionAttributeNames={"#request_status": "Status"},
                    ExpressionAttributeValues={":s": "Failed"},
                    ReturnValues="NONE",
                )
            except ClientError as err:
                logger.error(
                    "Couldn't update status %s in table %s. Here's why: %s: %s",
                    "Completed",
                    REQUESTS_TABLE_NAME,
                    err.response["Error"]["Code"],
                    err.response["Error"]["Message"],
                )
            continue

        try:
            ddb_response = requests_table.update_item(
                Key={"Id": request_id},
                UpdateExpression="set #request_status=:s",
                ExpressionAttributeNames={"#request_status": "Status"},
                ExpressionAttributeValues={":s": "Completed"},
                ReturnValues="NONE",
            )
        except ClientError as err:
            logger.error(
                "Couldn't update status %s in table %s. Here's why: %s: %s",
                "Completed",
                REQUESTS_TABLE_NAME,
                err.response["Error"]["Code"],
                err.response["Error"]["Message"],
            )

        logger.info("Updated status in requests table")

        #### Add to Allow List Table ####
        try:
            ddb_response = allow_list_table.put_item(Item=ip_address_obj)
            logger.info("Successfully added to allow list table...")
        except ClientError as err:
            logger.error(
                "Couldn't add ip object %s to table %s. Here's why: %s: %s",
                ip_address_obj,
                ALLOW_LIST_TABLE_NAME,
                err.response["Error"]["Code"],
                err.response["Error"]["Message"],
            )

    logger.info("Ended ip-retention-handler...")
    return
