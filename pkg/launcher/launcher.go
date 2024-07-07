package launcher

import (
	"gh-webhook/pkg/config"
	"gh-webhook/pkg/model"
)

const (
	NoAuth    = "none"
	BasicAuth = "basic"
	TokenAuth = "token"
)

const (
	Http    = "http"
	Jenkins = "jenkins"
)

var SupportedReceiverType = []string{Http, Jenkins}

var SupportedAuthType = []string{NoAuth, BasicAuth, TokenAuth}

type GHWebhookReceiverLauncher interface {
	Launch(routineId int32, config *config.Config, re model.GHWebHookReceiver, event model.GHWebHookEvent) error
}
