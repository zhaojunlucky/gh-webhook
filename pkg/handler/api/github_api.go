package api

import (
	"errors"
	"gh-webhook/pkg/core"
	"gh-webhook/pkg/model"
	"github.com/dranikpg/dto-mapper"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type GitHubCreateDTO struct {
	Web  string `json:"web" binding:"required"`
	API  string `json:"api" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type GitHubUpdateDTO struct {
	Web  string `json:"web" `
	API  string `json:"api" `
	Name string `json:"name" `
}

type GitHubSearchDTO struct {
	ID        uint      `json:"id" rsql:"id,filter,sort"`
	CreatedAt time.Time `json:"createdAt" `
	UpdatedAt time.Time `json:"updatedAt" `
	Web       string    `json:"web" rsql:"web,filter,sort"`
	API       string    `json:"api" rsql:"api,filter,sort"`
	Name      string    `json:"name" rsql:"name,filter,sort"`
}

type GitHubAPIHandler struct {
	db *gorm.DB
}

func (h *GitHubAPIHandler) Register(c *core.GHPRContext) error {
	h.db = c.Db
	return nil
}

func (h *GitHubAPIHandler) Post(c *gin.Context) {
	var ghCreateDTO GitHubCreateDTO

	if err := c.ShouldBindJSON(&ghCreateDTO); err != nil {
		log.Errorf("failed to bind json: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}
	github := model.GitHub{
		Web:  ghCreateDTO.Web,
		API:  ghCreateDTO.API,
		Name: ghCreateDTO.Name,
	}
	db := h.db.Save(&github)
	if db.Error != nil {
		log.Errorf("failed to save github: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}

	c.JSON(http.StatusCreated, github)
}

func (h *GitHubAPIHandler) Get(c *gin.Context) {
	id, err := core.UIntParam(c, "id")
	if err != nil {
		log.Errorf("failed to convert id: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}
	github := model.GitHub{}
	db := h.db.First(&github, "id = ?", id)
	if db.Error != nil {
		log.Errorf("failed to find github: %v", db.Error)
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, model.NewErrorMsgDTO(http.StatusText(http.StatusNotFound)))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(err))
		return
	}
	mapper := dto.Mapper{}
	to := GitHubSearchDTO{}
	err = mapper.Map(&to, github)
	if err != nil {
		log.Errorf("failed to convert id: %v", err)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(err))
		return
	}

	c.JSON(http.StatusOK, to)
}

func (h *GitHubAPIHandler) Update(c *gin.Context) {
	id, err := core.UIntParam(c, "id")
	if err != nil {
		log.Errorf("failed to convert id: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}
	var ghUpdateDTO GitHubUpdateDTO
	if err = c.ShouldBindJSON(&ghUpdateDTO); err != nil {
		log.Errorf("failed to bind json: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}

	if len(ghUpdateDTO.API) <= 0 && len(ghUpdateDTO.Web) <= 0 && len(ghUpdateDTO.Name) <= 0 {
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTO("no field to update"))
		return
	}
	github := model.GitHub{}
	db := h.db.First(&github, "id = ?", id)
	if db.Error != nil {
		log.Errorf("failed to find github: %v", db.Error)
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, model.NewErrorMsgDTO(http.StatusText(http.StatusNotFound)))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(err))
		return
	}
	if len(ghUpdateDTO.API) > 0 {
		github.API = ghUpdateDTO.API
	}
	if len(ghUpdateDTO.Web) > 0 {
		github.Web = ghUpdateDTO.Web
	}
	if len(ghUpdateDTO.Name) > 0 {
		github.Name = ghUpdateDTO.Name
	}
	db = h.db.Save(&github)
	if db.Error != nil {
		log.Errorf("failed to save github: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}
	c.JSON(http.StatusOK, model.NewIDResponse(github.ID))
}

func (h *GitHubAPIHandler) Delete(c *gin.Context) {
	id, err := core.UIntParam(c, "id")
	if err != nil {
		log.Errorf("failed to convert id: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}
	db := h.db.Delete(&model.GitHub{}, "id = ?", id)
	if db.Error != nil {
		log.Errorf("failed to delete github: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}
	if db.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, model.NewErrorMsgDTO(http.StatusText(http.StatusNotFound)))
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *GitHubAPIHandler) List(c *gin.Context) {
	page := core.ParsePagination(c)

	rsqlQuery := core.NewRSQLHelper()
	err := rsqlQuery.ParseFilter(GitHubSearchDTO{}, c)
	if err != nil {
		log.Errorf("failed to parse filter: %v", err)
		c.JSON(http.StatusBadRequest, model.NewErrorMsgDTOFromErr(err))
		return
	}
	var githubs []model.GitHub
	db := h.db.Order(rsqlQuery.SortSQL).Where(rsqlQuery.FilterSQL, rsqlQuery.Arguments...).
		Limit(page.Size).Offset((page.Page - 1) * page.Size).Find(&githubs)

	if db.Error != nil {
		log.Errorf("failed to find githubs: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}
	var githubDTOs []GitHubSearchDTO
	mapper := dto.Mapper{}
	err = mapper.Map(&githubDTOs, githubs)
	if err != nil {
		log.Errorf("failed to find githubs: %v", db.Error)
		c.JSON(http.StatusUnprocessableEntity, model.NewErrorMsgDTOFromErr(db.Error))
		return
	}

	c.JSON(http.StatusOK, githubs)
}
