package model

import (
	"fmt"
	"strings"
)

const (
	BasicAuth = "basic"
	TokenAuth = "token"
	NoneAuth  = "none"
	HTTP      = "http"
	Jenkins   = "jenkins"
)

type GHWebhookReceiverConfig struct {
	Type      string // http or jenkins
	URL       string
	Auth      string
	Username  string
	Password  string
	Parameter string // optional
}

func (c *GHWebhookReceiverConfig) IsValid() error {
	if c.Auth != BasicAuth && c.Type != TokenAuth && c.Auth != NoneAuth {
		return fmt.Errorf("invalid auth type %s", c.Type)
	}

	if c.Type != HTTP && c.Type != Jenkins {
		return fmt.Errorf("invalid receiver type %s", c.Type)
	}

	if c.Type == Jenkins && strings.Trim(c.Parameter, " ") == "" {
		return fmt.Errorf("invalid parameter")
	}

	if c.Auth != NoneAuth && (c.Username == "" || c.Password == "") {
		return fmt.Errorf("username/token header or password/token value is empty")
	}

	return nil
}

/**
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
