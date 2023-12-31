# This Makefile contains shortcuts

# generate go code based on the proto files
genproto: 
	rm -rf protogen/*
	protoc -I=./proto/ --go_out=protogen --go-grpc_out=protogen  proto/*.proto

# deletes protogen and the generated Go code
clean:
	rm -rf protogen/*

# builds all images
build-all:
	docker build -f ./docker/sqs_service.Dockerfile . -t alvinlucillo/sqsservice
	docker build -f ./docker/sqs_client.Dockerfile . -t alvinlucillo/sqsclient

# creates Docker image for the sqsservice
build-sqsservice:
	docker build -f ./docker/sqs_service.Dockerfile . -t alvinlucillo/sqsservice

# creates Docker image for the sqsclient
build-client:
	docker build -f ./docker/sqs_client.Dockerfile . -t alvinlucillo/sqsclient

# loads images to minikube local cluster
# when k8s deployment looks for the images, it won't check in Docker Hub since images are already in the cluster
load-all-images:
	minikube image load alvinlucillo/sqsclient 
	minikube image load alvinlucillo/sqsservice

load-sqsservice-image:
	minikube image load alvinlucillo/sqsservice

# lists images in the minikube local cluster
list-images:
	minikube image ls --format table

# deploy to k8s local cluster
deploy-local:
	kubectl apply -f ./kubernetes/.

# delete resources from k8s local cluster
delete-local:
	kubectl delete -f ./kubernetes/.

# create secret
create-secret:
	kubectl create secret generic sqsserviceapp-secret --from-literal=APP_AWS_ACCESS_KEY_ID=$(APP_AWS_ACCESS_KEY_ID) --from-literal=APP_AWS_SECRET_ACCESS_KEY=$(APP_AWS_SECRET_ACCESS_KEY)

# delete secret
delete-secret:
	kubectl delete secret sqsserviceapp-secret

# describe secret
desc-secret:
	kubectl get secret sqsserviceapp-secret -o yaml