package api

import (
	"errors"
	"fmt"
	"gh-webhook/pkg/core"
	"gh-webhook/pkg/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

// GHWebhookReceiverAPIHandler path: gh-webhook-receiver
type GHWebhookReceiverAPIHandler struct {
	db *gorm.DB
}

type GHWebhookReceiverConfigCreateDTO struct {
	Type   string            `json:"Type" binding:"required"`
	Config map[string]string `json:"config" binding:"required"`
}

type GHWebhookReceiverCreateDTO struct {
	Name           string                           `json:"name" binding:"required"`
	GitHubId       uint                             `json:"githubId" binding:"required"`
	ReceiverConfig GHWebhookReceiverConfigCreateDTO `json:"receiverConfig" binding:"required"`
}

func (h *GHWebhookReceiverAPIHandler) Register(c *core.GHPRContext) error {
	h.db = c.Db
	c.Gin.POST(fmt.Sprintf("%s/gh-webhook-receiver/", c.Cfg.APIPrefix), h.Post)
	c.Gin.PATCH(fmt.Sprintf("%s/gh-webhook-receiver/", c.Cfg.APIPrefix), h.Post)
	c.Gin.DELETE(fmt.Sprintf("%s/gh-webhook-receiver/", c.Cfg.APIPrefix), h.Post)
	c.Gin.GET(fmt.Sprintf("%s/gh-webhook-receiver/", c.Cfg.APIPrefix), h.Post)
	c.Gin.GET(fmt.Sprintf("%s/gh-webhook-receiver/", c.Cfg.APIPrefix), h.Post)
	c.Gin.DELETE(fmt.Sprintf("%s/gh-webhook-receiver/", c.Cfg.APIPrefix), h.Post)
	return nil
}

// Post create a new webhook receiver
func (h *GHWebhookReceiverAPIHandler) Post(c *gin.Context) {
	var createDTO GHWebhookReceiverCreateDTO

	if err := c.ShouldBindJSON(&createDTO); err != nil {
		log.Errorf("failed to bind json: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}
	var github model.GitHub
	dbErr := h.db.Find(&github, "id = ?", createDTO.GitHubId)
	if dbErr.Error != nil {
		log.Errorf("failed to find github: %v", dbErr.Error)
		if errors.Is(dbErr.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, model.NewErrorMsgDTO(http.StatusText(http.StatusBadRequest)))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(dbErr.Error))
		return
	}
	receiver := model.GHWebhookReceiver{
		Name:     createDTO.Name,
		GitHubId: createDTO.GitHubId,
		ReceiverConfig: model.GHWebhookReceiverConfig{
			Type:   createDTO.ReceiverConfig.Type,
			Config: createDTO.ReceiverConfig.Config,
		},
		Subscribes: nil,
	}
	db := h.db.Save(&receiver)
	if db.Error != nil {
		log.Errorf("failed to save webhook receiver: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}
	c.JSON(http.StatusCreated, receiver)
}
