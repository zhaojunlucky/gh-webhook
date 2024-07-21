package launcher

import (
	"gh-webhook/pkg/config"
	"gh-webhook/pkg/model"
)

var SupportedReceiverType = []string{model.HTTP, model.Jenkins}

var SupportedAuthType = []string{model.NoneAuth, model.BasicAuth, model.TokenAuth}

type GHWebhookReceiverLauncher interface {
	Launch(routineId int32, config *config.Config, re model.GHWebhookReceiver, event model.GHWebhookEvent,
		receiverDeliver model.GHWebhookEventReceiverDeliver) error
}
