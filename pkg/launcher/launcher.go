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
	Local   = "local"
	Jenkins = "jenkins"
)

var SupportedReceiverType = []string{Http, Local, Jenkins}

var SupportedAuthType = []string{NoAuth, BasicAuth, TokenAuth}

type GHWebhookReceiverLauncher interface {
	Launch(routineId int32, config config.Config, re model.GHWebHookReceiver, event model.GHWebHookEvent) error
}
