package model

import "gorm.io/gorm"

// GHWebHookEvent store webhook payload
type GHWebHookEvent struct {
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
