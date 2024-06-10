package core

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
	"time"
)

type TestRSQLDTO struct {
	ID        uint      `json:"id" rsql:"id,filter,sort"`
	CreatedAt time.Time `json:"createdAt" `
	UpdatedAt time.Time `json:"updatedAt" `
	Web       string    `json:"web" rsql:"web,filter,sort"`
	API       string    `json:"api" rsql:"api,filter,sort"`
	Name      string    `json:"name" rsql:"name,filter,sort"`
}

func Test_ParseFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = []gin.Param{gin.Param{Key: "query", Value: "web='github.com'"}}

	helper := NewRSQLHelper()

	helper.ParseFilter(TestRSQLDTO{}, c)

}
