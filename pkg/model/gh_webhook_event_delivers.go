package model

import "gorm.io/gorm"

type GHWebHookEventReceiverDeliver struct {
	GHWebHookReceiverId uint
	Delivered           bool
	Error               string
}

type GHWebhookEventDelivers struct {
	gorm.Model
	GHWebHookEventId   uint
	GHWebHookEvent     GHWebHookEvent
	GHWebHookReceivers []GHWebHookEventReceiverDeliver `gorm:"serializer:json"`
	Error              string
	Delivered          bool
}
