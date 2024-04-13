package model

import "gorm.io/gorm"

func Init(db *gorm.DB) error {
	err := db.AutoMigrate(&GitHub{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&GHWebHookReceiver{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&GHWebHookSubscribe{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&GHWebHookEvent{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&GHWebhookReceiverConfig{})
	if err != nil {
		return err
	}
	return nil
}
