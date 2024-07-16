package core

import (
	"github.com/gin-gonic/gin"
	"net/http"
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

func Test_ParseFilterInvalidField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "?filter=name==\"jun\";age==37", nil)

	helper := NewRSQLHelper()

	err := helper.ParseFilter(TestRSQLDTO{}, c)
	if err == nil {
		t.Fatal("should be failed")
	}

	if err.Error() != "field age is not allowed to filter" {
		t.Fatal(err)
	}

}
