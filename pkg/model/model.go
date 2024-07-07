package model

import "gorm.io/gorm"

func Init(db *gorm.DB) error {
	err := db.AutoMigrate(&GitHub{}, &GHWebHookReceiver{}, &GHWebHookEvent{}, &GHWebhookReceiverConfig{})
	if err != nil {
		return err
	}
	return nil
}
