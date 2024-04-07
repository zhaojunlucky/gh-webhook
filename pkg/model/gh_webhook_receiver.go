package model

import "gorm.io/gorm"

type GHWebHookReceiver struct {
	gorm.Model
	Name           string
	GitHubId       uint
	GitHub         GitHub
	ReceiverConfig GHWebhookReceiverConfig `gorm:"serializer:json"` // credentials and etc.
	Subscribes     []GHWebHookSubscribe    `gorm:"serializer:json"`
}
