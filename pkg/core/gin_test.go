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
	"time"
)

func TestGetModelNotFound(t *testing.T) {
	mockDb, mock, _ := sqlmock.New()
	dialector := sqlite.Dialector{
		DriverName: "sqlite",
		Conn:       mockDb,
	}
	mock.ExpectQuery("select sqlite_version()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("3.8.10"))

	db, _ := gorm.Open(dialector, &gorm.Config{})

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "?sort=name,asc;id,desc", nil)
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at",
		"deleted_at", "web", "api", "name"}))
	if GetModel(c, db, &model.GitHub{}, "id=?", 1) {
		t.Fatal("should be failed")
	}

	if w.Code != 404 {
		t.Fatal("should be 404")
	}

}

func TestGetModelSuccess(t *testing.T) {
	mockDb, mock, _ := sqlmock.New()
	dialector := sqlite.Dialector{
		DriverName: "sqlite",
		Conn:       mockDb,
	}
	mock.ExpectQuery("select sqlite_version()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("3.8.10"))

	db, _ := gorm.Open(dialector, &gorm.Config{})
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "?sort=name,asc;id,desc", nil)
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at",
		"deleted_at", "web", "api", "name"}).AddRow(1, time.Now(), time.Now(), nil, "github.com", "https://api.github.com", "jun"))
	gh := model.GitHub{}
	if !GetModel(c, db, &gh, "id=1") {
		t.Fatal("should be successful")
	}

	if gh.ID != 1 {
		t.Fatal("id should be equal 1")
	}

	if w.Code != 200 {
		t.Fatal("status code should be 200")
	}

	if w.Body.Available() != 0 {
		t.Fatal("body should be nil")
	}
}

func Test_GetPathVarUIntSuccessful(t *testing.T) {

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	id := GetPathVarUInt(c, "id")
	if id == nil {
		t.Fatal("should be successful")
	}
	if *id != 1 {
		t.Fatal("id should be equal 1")
	}
}

func Test_GetPathVarUIntFailed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "", nil)
	id := GetPathVarUInt(c, "id")
	if id != nil {
		t.Fatal("should be failed")
	}
	c.Params = []gin.Param{{Key: "id", Value: "sdsd"}}

	id = GetPathVarUInt(c, "id")
	if id != nil {
		t.Fatal("should be failed")
	}
}

func Test_SearchModel(t *testing.T) {
	mockDb, mock, _ := sqlmock.New()
	dialector := sqlite.Dialector{
		DriverName: "sqlite",
		Conn:       mockDb,
	}
	mock.ExpectQuery("select sqlite_version()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("3.8.10"))

	db, _ := gorm.Open(dialector, &gorm.Config{})
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "?sort=name,asc;id,desc", nil)
	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at",
		"deleted_at", "web", "api", "name"}).AddRow(1, time.Now(), time.Now(), nil, "github.com", "https://api.github.com", "jun"))
	var gh []model.GitHub

	type GitHubSearchDTO struct {
		ID        uint      `json:"id" rsql:"id,filter,sort"`
		CreatedAt time.Time `json:"createdAt" `
		UpdatedAt time.Time `json:"updatedAt" `
		Web       string    `json:"web" rsql:"web,filter,sort"`
		API       string    `json:"api" rsql:"api,filter,sort"`
		Name      string    `json:"name" rsql:"name,filter,sort"`
	}
	if !SearchModel(c, db, GitHubSearchDTO{}, &gh) {
		t.Fatal("should be successful")
	}

	if len(gh) != 1 || gh[0].ID != 1 {
		t.Fatal("should successful")
	}
}
