package types

type ServerEnvironment struct {
	Region    string `required:"true" default:"us-east-1"`
	QueueName string `required:"true" default:"sqs-sample-1"`
	Profile   string `required:"true" default:"default"`
	Port      int    `required:"true" default:"50051"`
}
