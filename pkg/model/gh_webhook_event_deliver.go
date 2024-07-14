package model

import "gorm.io/gorm"

type GHWebhookEventReceiverDeliver struct {
	gorm.Model
	GHWebhookReceiverId     uint
	GHWebhookEventDeliverID uint
	GHWebhookEventDeliver   GHWebhookEventDeliver
	Delivered               bool
	Error                   string
	Ack                     string
}

type GHWebhookEventDeliver struct {
	gorm.Model
	GHWebhookEventId   uint
	GHWebhookEvent     GHWebhookEvent
	GHWebhookReceivers []GHWebhookEventReceiverDeliver
	Error              string
	Delivered          bool
}
