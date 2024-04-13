package core

import (
	"gh-webhook/pkg/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GHPRContext struct {
	Gin *gin.Engine
	Db  *gorm.DB
	Cfg *config.Config
}
