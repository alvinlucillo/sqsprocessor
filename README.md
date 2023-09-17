# SQS Processor

- [SQS Processor](#sqs-processor)
  - [Tools/tech used 🛠️](#toolstech-used-️)
  - [Prerequisites 🖥️](#prerequisites-️)
  - [Setting up your environment 🚀](#setting-up-your-environment-)
    - [1. Creating SQS queue and IAM roles, policies and group](#1-creating-sqs-queue-and-iam-roles-policies-and-group)
    - [2. Deploying the project to a local cluster via minikube](#2-deploying-the-project-to-a-local-cluster-via-minikube)
    - [3. (Optional) Creating your own images](#3-optional-creating-your-own-images)
    - [4. (Optional) Deploying the project to EKS](#4-optional-deploying-the-project-to-eks)
  - [The process in action! 💪](#the-process-in-action-)

SQS Processor demonstrates the sidecar pattern, which is used to decouple processes like business logic or utility functionalities from the parent container to additional containers. SQS (Simple Queue Service) is a message queuing service from AWS providing facilities to manage queues.

In the diagram below, there are two containers:

- `sqsclient`
  - consumes the services provided by the sqsservice container
  - business logic resides here to process SQS messages
- `sqsservice`
  - the sidecar container provides endpoints to consume and delete SQS messages

In summary, `sqsclient` periodically looks for any messages from the queue via the endpoint exposed by sidecar container, sqsservice. The received messages are then deleted by `sqsclient` via another endpoint from sqsservice.

The diagram shows the system is deployed into a Kubernetes cluster using AWS services, but this project can also be tested in local cluster using minikube.

![Alt text](diagram.png?raw=true "Title")

The story is published here: https://medium.com/nullifying-the-null/sidecar-pattern-with-go-sqs-and-k8s-926d93166d0c

## Tools/tech used 🛠️

- Go with gRPC
- Kubernetes and Docker
- Terraform
- AWS

## Prerequisites 🖥️

- AWS account
- AWS CLI: https://aws.amazon.com/cli/
- Go: https://go.dev/doc/install
- Docker: https://docs.docker.com/engine/install/
- Terraform: https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli
- minikube: https://minikube.sigs.k8s.io/docs/start/
- kubectl: https://kubernetes.io/docs/tasks/tools/

## Setting up your environment 🚀

The setup is summarized into the following. This will let you just run existing containers and test them on your local machine.

1. Creating SQS queue and IAM roles, policies and group
2. Deploying the project to a local cluster via minikube

These are the optional steps if you want to take it a step further:

3. Creating your own images
4. Deploying the project to EKS

### 1. Creating SQS queue and IAM roles, policies and group

1. Set up AWS CLI to ensure your machine can connect to AWS: https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html

   - If you get a response from this command with your username, then you can connect to AWS: `aws sts get-caller-identity`

2. Set up terraform
   - Note: This will create SQS and IAM resources. The former is the minimum required but the latter is needed if you want to deploy the project to AWS.
   - `cd terraform`
   - `make init` to initialize your environment
   - `make plan` to to create a terraform plan
     - It should show this: `Plan: 11 to add, 0 to change, 0 to destroy.` This indicates what are the objects and the actions to be applied to them
   - `make apply` to perform the actions from the plan
     - It should show this: `Apply complete! Resources: 11 added, 0 changed, 0 destroyed.`
     - If you encounter permission problems, check the policies attached to the user you used in setting up your AWS CLI. For this project, it should have full access to SQS, S3, and IAM.

### 2. Deploying the project to a local cluster via minikube

1. Create Access Key ID and Secret Key Access Key in AWS Console
2. Export the keys. This will be used by the next step
   - `export APP_AWS_ACCESS_KEY_ID=your_access_key_here`
   - `export APP_AWS_SECRET_ACCESS_KEY=your_secret_access_key_here`
3. In the project root directory, run `make create-secret`
   - `make desc-secret` to check if secret is created
4. Run `make deploy-local`
   - `kubectl get all` to check all resources

### 3. (Optional) Creating your own images

1. If you make any changes to .proto files, run `make genproto`
2. If you want to create your own images
   - Run `make build-all`
   - Tag the newly created images to use your own Docker Hub username
     - `docker tag alvinlucillo/sqsservice your_username/sqsservice`
     - `docker tag alvinlucillo/sqsclient your_username/sqsclient`
   - Push to your Docker Hub account:
     - `docker push your_username/sqsservice:latest`
     - `docker push your_username/sqsclient:latest`

### 4. (Optional) Deploying the project to EKS

1. Create a VPC for EKS
   - Run
     `aws cloudformation create-stack \
--region us-east-1 \
--stack-name sqsservice-stack \
--template-url https://s3.us-west-2.amazonaws.com/amazon-eks/cloudformation/2020-10-29/amazon-eks-vpc-private-subnets.yaml`
2. Create your EKS cluster (via AWS Console)
   - Go to EKS and choose Add cluster > Create
   - Enter a name (e.g., sqsservice-cluster)
   - Choose the service role `sqsservice-cluster-role` created by terraform step
   - Choose `sqsservice-stack-VPC` created by an earlier step, then click Create to create the cluster. Wait until the status is Active.
   - In the cluster, under Compute tab, click Add Fargate Profile then enter a name (e.g., sqsservice-fargate-profile)
   - Choose the service role `sqsservice-fargate-pod-execn-role` created by terraform step
   - Enter `default` namespace, then click Create to create the Fargate profile. Wait until the status is Active.
3. If the user you configured with AWS CLI is the same user you used in creating the EKS cluster, you can skip this step. Otherwise, you need to set up the cluster to enable a new user. As a best practice, a non-root user should be used in the AWS CLI, that's why we have this step to enable other users (e.g., developer, non-admin)
   - Add the new user to the group `sqsservice-group-developer` created by terraform step
   - Access CloudShell on AWS Console. It might take a few minutes to initialize it. Set up your kubeconfig there: `aws eks update-kubeconfig --region your_region --name your_cluster_name`. Replace the region and cluster name (e.g., `us-east-1` and `sqsservice-cluster`). Cluster name should be the same name you used in the earlier step.
   - Install eksctl: https://eksctl.io/introduction/
   - Add an identity mapping to register your role to cluster. Any users that are in the group created by terraform can assume this role to access the cluster. Replace the account number with your own from AWS console:  `eksctl create iamidentitymapping --cluster  sqsservice-cluster --region=us-east-1 --arn arn:aws:iam::your_account_number:role/sqsservice-eks-readonly-role --group system:masters --username admin`
   - Check if the identity mapping is created: `eksctl get iamidentitymapping --cluster sqsservice-cluster`
4. Update your kubeconfig
   - If you skipped step #3 because the user you used for EKS cluster creation in AWS Console and for configuring AWS CLI, use this: `aws eks update-kubeconfig --name sqsservice-cluster --region us-east-1`
   - Otherwise, follow the steps below:
     - The user's access key id and secret access keys are in `~/.aws/credentials`:
     ```
     [devprofile]
     aws_access_key_id = ...
     aws_secret_access_key = ...
     ```
      - The role is defined in `~/.aws/config`
      ```
      [profile eksroleprofile]
      role_arn = arn:aws:iam::your_account_number:role/sqsservice-eks-readonly-role
      source_profile = devprofile
      ```
      - Update your kubeconfig: `aws eks update-kubeconfig --name sqsservice-cluster --profile eksroleprofile --region us-east-1`
    - Ensure that the current context is set to EKS: `kubectl config current-context `
    - Check if you can list all resources: `kubectl get all`
5. Since the context is now set to EKS, apply the same steps in [2. Deploying the project to a local cluster via minikube](#2-deploying-the-project-to-a-local-cluster-via-minikube)

## The process in action! 💪
1. Stream the logs for the two containers:
    - sqsservice: `kubectl logs pod/pod_name -c sqsservice -f`
    - sqsclient: `kubectl logs pod/pod_name -c sqsclient -f`
2. While there's no message in the queue, you'll see the following
    - sqsclient - this is configured to continuously poll for messages from sqsservice
        ```
        {"level":"info","caller":"/app/cmd/client/main.go:16","time":"2023-09-17T10:12:58Z","message":"Client starting"}
        {"level":"info","package":"client","function":"Run","time":"2023-09-17T10:12:58Z","message":"Polling count: 1"}{"level":"info","package":"client","function":"Run","time":"2023-09-17T10:13:03Z","message":"Received 0 message(s)"}
        {"level":"info","package":"client","function":"Run","time":"2023-09-17T10:13:08Z","message":"Polling count: 2"}
         ```
    - sqsservice - this is also configured to continuously poll for messages based on parameters set to the service
        ```
        {"level":"info","caller":"/app/cmd/sqsservice/main.go:18","time":"2023-09-17T10:12:55Z","message":"Server starting"}
        {"level":"debug","function":"ReceiveMessage","time":"2023-09-17T10:12:58Z","message":"Received input: visibility_timeout:5  wait_time:5  maximum_number_of_messages:5"}
        {"level":"debug","function":"ReceiveMessage","time":"2023-09-17T10:13:03Z","message":"Returned output: []"}
        ```
3. Create an SQS message
    - `aws sqs send-message --queue-url https://sqs.us-east-1.amazonaws.com/your-aws_account_no/sqs-sample-1  --message-body "hello"` - replace `your-aws_account_no` with your AWS account number

4. SQSService will pick up the new message
   ```
   {"level":"debug","function":"ReceiveMessage","time":"2023-09-17T10:15:14Z","message":"Returned output: [messageID:\"AQEBo4QCIVXuMaMZew==\"  messageBody:\"hello\"]"}
   ```

5. SQSClient will receive the message from the SQSService then delete it
   ```
    {"level":"info","package":"client","function":"Run","time":"2023-09-17T10:15:14Z","message":"Received 1 message(s)"}
    {"level":"info","package":"client","function":"Run","time":"2023-09-17T10:15:14Z","message":"Deleting message messageID:\"AQEBo4QCIVXuMaMZew==\" messageBody:\"hello\""}
    {"level":"info","package":"client","function":"Run","time":"2023-09-17T10:15:14Z","message":"Message deleted successfully: AQEBo4QCIVXuMaMZew==\"}
   ```