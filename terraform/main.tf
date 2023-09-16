# The empty block is also called partial configuration which
#  inform Terraform that backend configuration is defined dynamically.
#  This project uses -backend-configuration (see Makefile)
terraform {
  backend "s3" {}
}

provider "aws" {
  region = var.region
}

data "aws_caller_identity" "current" {}

# -- SQS Queues
resource "aws_sqs_queue" "sqs_queues" {
  for_each = var.sqs_queues

  name                       = each.value.name
  delay_seconds              = each.value.delay_seconds
  max_message_size           = each.value.max_message_size
  message_retention_seconds  = each.value.message_retention_seconds
  visibility_timeout_seconds = each.value.visibility_timeout_seconds
}

# -- Roles, permissions and policies

# EKS cluster role that the EKS will assume
resource "aws_iam_role" "sqsservice_cluster_role" {
  name               = "sqsservice-cluster-role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

# Attaches what the permissions the EKS will have via the indicated policy 
#  when it assumes the EKS cluster role
resource "aws_iam_role_policy_attachment" "sqsservice_cluster_role_cluster_policy" {
  role       = aws_iam_role.sqsservice_cluster_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
}

# Fargate execution role that Fargate will assume
resource "aws_iam_role" "sqsservice_fargate_pod_execn_role" {
  name               = "sqsservice-fargate-pod-execn-role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Condition": {
         "ArnLike": {
            "aws:SourceArn": "arn:aws:eks:${var.region}:${data.aws_caller_identity.current.account_id}:fargateprofile/${var.cluster_name}/*"
         }
      },
      "Principal": {
        "Service": "eks-fargate-pods.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

# Attaches what the permissions the Fargate will have via the indicated policy 
#  when it assumes the Fargate execution role
resource "aws_iam_role_policy_attachment" "sqsservice_fargate_pod_execn_role_policy" {
  role       = aws_iam_role.sqsservice_fargate_pod_execn_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSFargatePodExecutionRolePolicy"
}

# The policy that allows users to read EKS resources;
# Will be used by developers
resource "aws_iam_policy" "sqsservice_eks_readonly_policy" {
  name = "sqsservice-eks-readonly-policy"

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "EksReadonly",
            "Effect": "Allow",
            "Action": [
                "eks:ListNodegroups",
                "eks:DescribeFargateProfile",
                "eks:ListTagsForResource",
                "eks:ListAddons",
                "eks:DescribeAddon",
                "eks:ListFargateProfiles",
                "eks:DescribeNodegroup",
                "eks:DescribeIdentityProviderConfig",
                "eks:ListUpdates",
                "eks:DescribeUpdate",
                "eks:AccessKubernetesApi",
                "eks:DescribeCluster",
                "eks:ListIdentityProviderConfigs"
            ],
            "Resource": "arn:aws:eks:${var.region}:${data.aws_caller_identity.current.account_id}:cluster/${var.cluster_name}"
        }
    ]
}
EOF
}

# The role that the developers will assume to read EKS resources
resource "aws_iam_role" "sqsservice_eks_readonly_role" {
  name               = "sqsservice-eks-readonly-role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

# Attaches the permissions of the EKS readonly policy created in this project
#  to the readonly role also created in the project
resource "aws_iam_role_policy_attachment" "sqsservice_eks_readonly_role_policy" {
  role       = aws_iam_role.sqsservice_eks_readonly_role.name
  policy_arn = aws_iam_policy.sqsservice_eks_readonly_policy.arn

  # ensures the dependent resources are created first before creating the attachment
  depends_on = [aws_iam_policy.sqsservice_eks_readonly_policy, aws_iam_role.sqsservice_eks_readonly_role]
}

# The policy that allows entities to assume the EKS readonly role
resource "aws_iam_policy" "sqsservice_eks_readonly_assume_role_policy" {
  name = "sqsservice-eks-readonly-assume-role-policy"

  policy     = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "SqsserviceEksReadonlyAssumeRolePolicy",
            "Effect": "Allow",
            "Action": [
                "sts:AssumeRole"
            ],
            "Resource": [
                "${aws_iam_role.sqsservice_eks_readonly_role.arn}"
            ]
        }
    ]
}
EOF
  depends_on = [aws_iam_role.sqsservice_eks_readonly_role]
}

# Creates a group for developers
resource "aws_iam_group" "sqsservice_group_developer" {
  name = "sqsservice-group-developer"
}

# Attaches the assume policy created in this project
#  to the group also created in this project;
#  this allows users in this group to assume the role,
#  thereby allowing to read EKS resources
resource "aws_iam_group_policy_attachment" "sqsservice_developer_group_eks_readonly_assume_policy" {
  policy_arn = aws_iam_policy.sqsservice_eks_readonly_assume_role_policy.arn
  group      = aws_iam_group.sqsservice_group_developer.name
  depends_on = [aws_iam_group.sqsservice_group_developer, aws_iam_policy.sqsservice_eks_readonly_assume_role_policy]
}
