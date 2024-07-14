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

type GHWebhookEventAPIHandler struct {
	db *gorm.DB
}

type GHWebhookEventSearchDTO struct {
	ID        uint              `json:"id" rsql:"id,filter,sort"`
	HookMeta  map[string]string `json:"hookMeta"` // Hook headers from GitHub
	Payload   string            `json:"payload"`  // raw payload from GitHub
	Event     string            `json:"event" rsql:"event,filter,sort"`
	Action    string            `json:"action" rsql:"action,filter,sort"`
	OrgRepo   string            `json:"orgRepo" rsql:"orgRepo,filter,sort"`
	HookId    string            `json:"hookId" rsql:"hookId,filter,sort"`
	PayloadId string            `json:"payloadId" rsql:"payloadId,filter,sort"`
	GitHubId  uint              `json:"githubId" rsql:"githubId,filter,sort"`
}

func (h *GHWebhookEventAPIHandler) Register(c *core.GHPRContext) error {
	c.Gin.GET(fmt.Sprintf("%s/gh-webhook-event/:id", c.Cfg.APIPrefix), h.Get)
	c.Gin.GET(fmt.Sprintf("%s/gh-webhook-event/", c.Cfg.APIPrefix), h.List)
	return nil
}

func (h *GHWebhookEventAPIHandler) Get(c *gin.Context) {
	id := core.GetPathVarUInt(c, "id")
	if id == nil {
		return
	}
	ghEvent := model.GHWebhookEvent{}

	if !core.GetModel(c, h.db, &ghEvent, "id = ?", *id) {
		return
	}

	mapper := dto.Mapper{}
	to := GHWebhookEventSearchDTO{}
	err := mapper.Map(&to, ghEvent)
	if err != nil {
		log.Errorf("failed to convert id: %v", err)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(err))
		return
	}
	c.JSON(http.StatusOK, to)

}

func (h *GHWebhookEventAPIHandler) List(c *gin.Context) {

	var subs []model.GHWebhookReceiver
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
