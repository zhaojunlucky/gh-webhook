package model

const (
	BasicAuth = "basic"
	TokenAuth = "token"
)

type GHWebhookReceiverConfig struct {
	Type   string                 // local or http or
	Config map[string]interface{} `gorm:"serializer:json"`
}

/**
For local receiver,
---
{
	app: curl
	payloadArg: --event
	args: ["-a", "a"]
}

For jenkins receiver,
---
{
	"url": "http://127.0.0.1:8080/job/aa/build",
	"auth": "basic or token"
	"username": "username",
	"password": "password or token",
	"parameter": "payload"
}

For http receiver,
---
{
	"url": "http://127.0.0.1:8080/job/aa/build",
	"auth": "basic or token"
	"username": "username or token header name",
	"password": "password or token",
}

*/
