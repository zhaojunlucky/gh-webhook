package api

import (
	"fmt"
	"gh-webhook/pkg/core"
	"gh-webhook/pkg/model"
	"github.com/dranikpg/dto-mapper"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

type GHWebhookEventReceiverDeliverAPIHandler struct {
	db *gorm.DB
}

type GHWebhookEventReceiverDeliverSearchDTO struct {
	ID                      uint   `json:"id" rsql:"id,filter,sort"`
	GHWebhookReceiverId     uint   `json:"GHWebhookReceiverId" rsql:"ghWebhookReceiverId,filter,sort"`
	GHWebhookEventDeliverID uint   `json:"ghWebhookEventDeliverId" rsql:"ghWebhookEventDeliverId,filter,sort"`
	Delivered               bool   `json:"delivered" rsql:"delivered,filter,sort"`
	Error                   string `json:"error" rsql:"error,filter,sort"`
	Ack                     string `json:"ack" rsql:"ack,filter,sort"`
}

type GHWebhookEventReceiverDeliverAckCreateDTO struct {
	Ack string `json:"ack" binding:"required"`
}

func (h *GHWebhookEventReceiverDeliverAPIHandler) Register(c *core.GHPRContext) error {
	h.db = c.Db
	c.Gin.GET(fmt.Sprintf("%s/gh-webhook-event-receiver-deliver", c.Cfg.APIPrefix), h.List)
	c.Gin.GET(fmt.Sprintf("%s/gh-webhook-event-receiver-deliver/:id", c.Cfg.APIPrefix), h.Get)
	c.Gin.PUT(fmt.Sprintf("%s/gh-webhook-event-receiver-deliver/:id/ack", c.Cfg.APIPrefix), h.Put)
	return nil
}

func (h *GHWebhookEventReceiverDeliverAPIHandler) List(c *gin.Context) {
	var subs []model.GHWebhookEventReceiverDeliver
	if !core.SearchModel(c, h.db, GHWebhookEventReceiverDeliverSearchDTO{}, &subs) {
		return
	}
	var receiverDTOs []GHWebhookEventReceiverDeliverSearchDTO
	mapper := dto.Mapper{}
	err := mapper.Map(&receiverDTOs, subs)
	if err != nil {
		log.Errorf("failed to map: %v", err)
		c.JSON(http.StatusInternalServerError, model.NewErrorMsgDTOFromErr(err))
		return
	}

	c.JSON(http.StatusOK, model.NewListResponse(receiverDTOs))
}

func (h *GHWebhookEventReceiverDeliverAPIHandler) Get(c *gin.Context) {
	id := core.GetPathVarUInt(c, "id")
	if id == nil {
		return
	}
	deliver := model.GHWebhookEventReceiverDeliver{}
	if !core.GetModel(c, h.db, &deliver, "id = ?", *id) {
		return
	}
	mapper := dto.Mapper{}
	to := GHWebhookEventReceiverDeliverSearchDTO{}
	err := mapper.Map(&to, deliver)
	if err != nil {
		log.Errorf("failed to map: %v", err)
		c.JSON(http.StatusInternalServerError, model.NewErrorMsgDTOFromErr(err))
		return
	}
	c.JSON(http.StatusOK, to)
}

func (h *GHWebhookEventReceiverDeliverAPIHandler) Put(c *gin.Context) {
	id := core.GetPathVarUInt(c, "id")
	if id == nil {
		return
	}

	ack := GHWebhookEventReceiverDeliverAckCreateDTO{}

	if err := c.ShouldBindJSON(&ack); err != nil {
		log.Errorf("failed to bind json: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}

	deliver := model.GHWebhookEventReceiverDeliver{}
	if !core.GetModel(c, h.db, &deliver, "id = ?", *id) {
		return
	}
	deliver.Ack = ack.Ack
	err := h.db.Save(&deliver).Error
	if err != nil {
		log.Errorf("failed to save: %v", err)
		c.JSON(http.StatusInternalServerError, model.NewErrorMsgDTOFromErr(err))
		return
	}
	c.JSON(http.StatusOK, nil)
}
