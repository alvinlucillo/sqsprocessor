package types

type Env struct {
	Region    string `def:"region,us-east-1"`
	QueueName string `def:"queue_name,sqs-sample-1"`
	Profile   string `def:"profile,default"`
	Port      int    `def:"port,50051"`
}

type ServerEnvironment struct {
	Region    string `required:"true" default:"us-east-1"`
	QueueName int    `required:"true" default:"sqs-sample-1"`
	Profile   int    `required:"true" default:"default"`
	Port      int    `required:"true" default:"50051"`
}
