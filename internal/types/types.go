package types

type Env struct {
	Region    string `def:"region,us-east-1"`
	QueueName string `def:"queue_name,sqs-sample-1"`
	Profile   string `def:"profile,default"`
	Port      int    `def:"port,50051"`
}
