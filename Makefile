genproto: 
	rm -rf protogen/*
	protoc -I=./proto/ --go_out=protogen --go-grpc_out=protogen  proto/*.proto
clean:
	rm -rf protogen/*