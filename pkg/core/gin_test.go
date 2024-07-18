package core

import (
	"gh-webhook/pkg/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetModel(t *testing.T) {
	mockDb, mock, _ := sqlmock.New()
	dialector := sqlite.Dialector{
		DriverName: "sqlite",
		Conn:       mockDb,
	}
	db, _ := gorm.Open(dialector, &gorm.Config{})

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "?sort=name,asc;id,desc", nil)
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("3.8.10"))
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at",
		"deleted_at", "web", "api", "name"}))
	if GetModel(c, db, &model.GitHub{}, "id=1") {
		t.Fatal("should be failed")
	}

}
