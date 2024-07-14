package model

import "gorm.io/gorm"

func Init(db *gorm.DB) error {
	err := db.AutoMigrate(&GitHub{}, &GHWebhookReceiver{}, &GHWebhookEvent{}, &GHWebHookSubscribe{},
		&GHWebhookEventDeliver{}, &GHWebhookEventReceiverDeliver{})
	if err != nil {
		return err
	}
	return nil
}
