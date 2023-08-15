# This Makefile contains shortcuts

# generate go code based on the proto files
genproto: 
	rm -rf protogen/*
	protoc -I=./proto/ --go_out=protogen --go-grpc_out=protogen  proto/*.proto

# deletes protogen and the generated Go code
clean:
	rm -rf protogen/*

# creates Docker image for the sqsservice
build-sqsservice:
	docker build -f ./docker/sqs_service.Dockerfile . -t alvinlucillo/sqsservice

# creates Docker image for the sqsservice
build-client:
	docker build -f ./docker/sqs_client.Dockerfile . -t alvinlucillo/sqsclient

# loads images to minikube local cluster
# when k8s deployment looks for the images, it won't check in Docker Hub since images are already in the cluster
load-images:
	minikube image load alvinlucillo/sqsclient 
	minikube image load alvinlucillo/sqsservice

# lists images in the minikube local cluster
list-images:
	minikube image ls --format table