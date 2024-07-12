package api

import (
	"errors"
	"fmt"
	"gh-webhook/pkg/core"
	"gh-webhook/pkg/model"
	"github.com/dranikpg/dto-mapper"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"time"
)

// GHWebhookReceiverAPIHandler path: gh-webhook-receiver
type GHWebhookReceiverAPIHandler struct {
	db *gorm.DB
}

type GHWebhookReceiverConfigCreateDTO struct {
	Type      string `json:"type" binding:"required"`
	URL       string `json:"url" binding:"required"`
	Auth      string `json:"auth" binding:"required"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Parameter string `json:"parameter" binding:"required"` // optional
}

type GHWebhookReceiverCreateDTO struct {
	Name           string                           `json:"name" binding:"required"`
	GitHubId       uint                             `json:"githubId" binding:"required"`
	ReceiverConfig GHWebhookReceiverConfigCreateDTO `json:"receiverConfig" binding:"required"`
}

type GHWebhookReceiverSearchDTO struct {
	ID             uint   `json:"id" rsql:"id,filter,sort"`
	Name           string `json:"name" rsql:"name,filter,sort"`
	GitHub         GitHubSearchDTO
	ReceiverConfig GHWebhookReceiverConfigSearchDTO `json:"config"`
	CreatedAt      time.Time                        `json:"createdAt" `
	UpdatedAt      time.Time                        `json:"updatedAt" `
}

type GHWebhookReceiverConfigSearchDTO struct {
	Type      string `json:"type"`
	URL       string `json:"url" `
	Auth      string `json:"auth" `
	Username  string `json:"username"`
	Password  string `json:"password"`
	Parameter string `json:"parameter" ` // optional
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
		GitHub:   github,
		ReceiverConfig: model.GHWebhookReceiverConfig{
			Type:      createDTO.ReceiverConfig.Type,
			URL:       createDTO.ReceiverConfig.URL,
			Auth:      createDTO.ReceiverConfig.Auth,
			Username:  createDTO.ReceiverConfig.Username,
			Password:  createDTO.ReceiverConfig.Password,
			Parameter: createDTO.ReceiverConfig.Parameter},
		Subscribes: nil,
	}

	if err := receiver.ReceiverConfig.IsValid(); err != nil {
		log.Errorf("invalid request: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}

	db := h.db.Save(&receiver)
	if db.Error != nil {
		log.Errorf("failed to save webhook receiver: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}
	c.JSON(http.StatusCreated, model.NewIDResponse(receiver.ID))
}

// Get get webhook receiver
func (h *GHWebhookReceiverAPIHandler) Get(c *gin.Context) {
	id, err := core.UIntParam(c, "id")
	if err != nil {
		log.Errorf("failed to convert id: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}
	receiver := model.GHWebhookReceiver{}
	db := h.db.First(&receiver, "id = ?", id)
	if db.Error != nil {
		log.Errorf("failed to find webhook receiver: %v", db.Error)
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, model.NewErrorMsgDTO(http.StatusText(http.StatusNotFound)))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}
	mapper := dto.Mapper{}
	to := GHWebhookReceiverSearchDTO{}
	err = mapper.Map(&to, receiver)
	c.JSON(http.StatusOK, receiver)
}

// Delete delete webhook receiver
func (h *GHWebhookReceiverAPIHandler) Delete(c *gin.Context) {
	id, err := core.UIntParam(c, "id")
	if err != nil {
		log.Errorf("failed to convert id: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}
	db := h.db.Delete(&model.GHWebhookReceiver{}, "id = ?", id)
	if db.Error != nil {
		log.Errorf("failed to delete webhook receiver: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}
	if db.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, model.NewErrorMsgDTO(http.StatusText(http.StatusNotFound)))
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// List list webhook receivers
func (h *GHWebhookReceiverAPIHandler) List(c *gin.Context) {
	page := core.ParsePagination(c)

	rsqlQuery := core.NewRSQLHelper()
	err := rsqlQuery.ParseFilter(GHWebhookReceiverSearchDTO{}, c)
	if err != nil {
		log.Errorf("failed to parse filter: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}
	var subs []model.GHWebhookReceiver
	db := h.db.Order(rsqlQuery.SortSQL).Where(rsqlQuery.FilterSQL, rsqlQuery.Arguments...).
		Limit(page.Size).Offset((page.Page - 1) * page.Size).Find(&subs)

	if db.Error != nil {
		log.Errorf("failed to find receivers: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}
	var receiverDTOs []GHWebhookReceiverSearchDTO
	mapper := dto.Mapper{}
	err = mapper.Map(&receiverDTOs, subs)
	if err != nil {
		log.Errorf("failed to find githubs: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}

	c.JSON(http.StatusOK, model.NewListResponse(receiverDTOs))
}
