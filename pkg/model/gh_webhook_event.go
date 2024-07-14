package model

import "gorm.io/gorm"

// GHWebhookEvent store webhook payload
type GHWebhookEvent struct {
	gorm.Model
	HookMeta  map[string]string `gorm:"serializer:json"` // Hook headers from GitHub
	Payload   string            // raw payload from GitHub
	Event     string
	Action    string
	OrgRepo   string
	HookId    string
	PayloadId string
	GitHubId  uint   // github id
	GitHub    GitHub // GitHub instance
}

type Queue chan GHWebhookEvent

var queue = make(Queue)

func GetQueue() Queue {
	return queue
}
