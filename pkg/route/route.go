package route

import (
	"gh-webhook/pkg/core"
	"gh-webhook/pkg/handler/api"
	"gh-webhook/pkg/handler/webhook"
)

var routers = []core.RouterRegister{
	&webhook.GHWebhookHandler{},
	&webhook.GHWebhookDeliverHandler{},
	&api.GHWebhookEventAPIHandler{},
	&api.GHWebhookReceiverAPIHandler{},
	&api.GHWebhookSubscribeAPIHandler{},
	&api.GitHubAPIHandler{},
}

func Init(ctx *core.GHPRContext) error {

	for _, r := range routers {
		if err := r.Register(ctx); err != nil {
			return err
		}
	}
	return nil
}
