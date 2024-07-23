package webhook

import (
	"database/sql/driver"
	"fmt"
	"gh-webhook/pkg/config"
	"gh-webhook/pkg/model"
	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"
)

func httpHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(200)
	fmt.Fprintf(writer, "Hello, client!")
	defer request.Body.Close()

	data, _ := io.ReadAll(request.Body)
	log.Infof("request body: %s", string(data))
}

func Test_handle(t *testing.T) {
	handler := GHWebhookDeliverHandler{}
	var routineId int32 = 1
	ghEvent := model.GHWebhookEvent{
		Model: gorm.Model{
			ID:        0,
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
			DeletedAt: gorm.DeletedAt{},
		},
		Payload:   `{"branch": "usr/main", "action": "push"}`,
		Event:     "push",
		Action:    "push",
		OrgRepo:   "zhaojunlucky/veda",
		HookId:    "test",
		PayloadId: "test",
		GitHubId:  1,
		GitHub: model.GitHub{
			Model: gorm.Model{
				ID:        1,
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
				DeletedAt: gorm.DeletedAt{},
			},
			Web:  "https://github.com",
			API:  "https://api.github.com/v3",
			Name: "github",
		},
	}

	mockDb, mock, _ := sqlmock.New()
	dialector := sqlite.Dialector{
		DriverName: "sqlite",
		Conn:       mockDb,
	}
	mock.ExpectQuery("select sqlite_version()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("3.8.10"))

	db, _ := gorm.Open(dialector, &gorm.Config{})

	handler.db = db
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `git_hubs`").WithArgs(PrepareArgs(7)...).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `gh_webhook_events`").WithArgs(PrepareArgs(11)...).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `gh_webhook_event_delivers`").WithArgs(PrepareArgs(6)...).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	ts := httptest.NewServer(http.HandlerFunc(httpHandler))
	defer ts.Close() // Close the server after the test

	receiverConfig := fmt.Sprintf(`
{
"type": "jenkins",
"auth": "none",
"parameter": "test",
"url": "%s"
}
`, ts.URL)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `gh_webhook_receivers`")).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at",
		"deleted_at", "name", "git_hub_id", "receiver_config"}).AddRow(1, time.Now(), time.Now(), nil, "gundam", 1, receiverConfig))

	filters := `{"action":{"positiveMatches":["push"],"negativeMatches":[]}}`
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `gh_web_hook_subscribes`")).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at",
		"deleted_at", "gh_webhook_receiver_id", "event", "filters"}).AddRow(1, time.Now(), time.Now(), nil, 1, "push", filters))

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `git_hubs`").WithArgs(PrepareArgs(7)...).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `gh_webhook_events`").WithArgs(PrepareArgs(12)...).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `gh_webhook_event_delivers`").WithArgs(PrepareArgs(7)...).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `gh_webhook_event_receiver_delivers`").WithArgs(PrepareArgs(8)...).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	handler.config = &config.Config{
		DBType:     "",
		DBDsn:      "",
		ListenAddr: "",
		APIUrl:     "",
		APIPrefix:  "",
	}

	handler.handle(routineId, ghEvent)
}

func PrepareArgs(count int) []driver.Value {
	var args []driver.Value
	for _ = range count {
		args = append(args, sqlmock.AnyArg())
	}
	return args
}
