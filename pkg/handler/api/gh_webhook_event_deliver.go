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

type GHWebhookEventDeliverAPIHandler struct {
	db *gorm.DB
}

type GHWebhookEventDeliverSearchDTO struct {
	ID                 uint                          `json:"id" rsql:"id,filter,sort"`
	GHWebhookEventId   uint                          `json:"ghWebhookEventId" rsql:"ghWebhookEventId,filter,sort"`
	GHWebhookReceivers []GHWebhookReceiversSearchDTO `json:"ghWebhookReceivers"`
	Error              string                        `json:"error" rsql:"error,filter,sort"`
	Delivered          bool                          `json:"delivered" rsql:"delivered,filter,sort"`
}

type GHWebhookReceiversSearchDTO struct {
	GHWebhookReceiverId uint   `json:"ghWebhookReceiverId"`
	Delivered           bool   `json:"delivered"`
	Error               string `json:"error"`
}

func (h *GHWebhookEventDeliverAPIHandler) Register(c *core.GHPRContext) error {
	h.db = c.Db
	c.Gin.GET(fmt.Sprintf("%s/gh-webhook-event-deliver", c.Cfg.APIPrefix), h.List)
	c.Gin.GET(fmt.Sprintf("%s/gh-webhook-event-deliver/:id", c.Cfg.APIPrefix), h.Get)
	return nil
}

func (h *GHWebhookEventDeliverAPIHandler) List(c *gin.Context) {
	var subs []model.GHWebhookEventDeliver
	if !core.SearchModel(c, h.db, GHWebhookEventSearchDTO{}, &subs) {
		return
	}
	var receiverDTOs []GHWebhookReceiverSearchDTO
	mapper := dto.Mapper{}
	err := mapper.Map(&receiverDTOs, subs)
	if err != nil {
		log.Errorf("failed to map: %v", err)
		c.JSON(http.StatusInternalServerError, model.NewErrorMsgDTOFromErr(err))
		return
	}

	c.JSON(http.StatusOK, model.NewListResponse(receiverDTOs))
}

func (h *GHWebhookEventDeliverAPIHandler) Get(c *gin.Context) {
	id := core.GetPathVarUInt(c, "id")
	if id == nil {
		return
	}
	deliver := model.GHWebhookEventDeliver{}
	if !core.GetModel(c, h.db, &deliver, "id = ?", *id) {
		return
	}
	mapper := dto.Mapper{}
	to := GHWebhookEventDeliverSearchDTO{}
	err := mapper.Map(&to, deliver)
	if err != nil {
		log.Errorf("failed to map: %v", err)
		c.JSON(http.StatusInternalServerError, model.NewErrorMsgDTOFromErr(err))
		return
	}
	c.JSON(http.StatusOK, to)
}
