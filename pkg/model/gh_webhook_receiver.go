package model

import "gorm.io/gorm"

type GHWebhookReceiver struct {
	gorm.Model
	Name           string
	GitHubId       uint
	GitHub         GitHub
	ReceiverConfig GHWebhookReceiverConfig `gorm:"serializer:json"` // credentials and etc.
	Subscribes     []GHWebHookSubscribe
}
