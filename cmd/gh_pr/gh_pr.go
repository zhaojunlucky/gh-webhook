package main

import (
	"flag"
	"fmt"
	"gh-webhook/pkg/config"
	"gh-webhook/pkg/core"
	"gh-webhook/pkg/model"
	"gh-webhook/pkg/route"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"os"
	"path"
	"runtime"
	"time"
)

func setupLog() error {
	if runtime.GOOS == "windows" {
		panic("Windows is currently not supported.")
	}
	logPath := "/var/log/gh_pr"
	fiInfo, err := os.Stat(logPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(logPath, 0755)
		if err != nil {
			panic(err)
		}
	} else if !fiInfo.IsDir() {
		panic(fmt.Sprintf("%s must be a directory.", logPath))
	}
	curTime := time.Now()
	nanoseconds := curTime.Nanosecond()
	formattedTime := curTime.UTC().Format("20060102150405")

	logFilePath := path.Join(logPath, fmt.Sprintf("gh_pr_%s_%d.log", formattedTime, nanoseconds))
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Info("Failed to log to file, using default stderr")
		return nil
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	return nil
}

func main() {
	if err := setupLog(); err != nil {
		panic(err)
	}

	var configFile = flag.String("c", "", "The config file")
	flag.Parse()

	cfg, err := config.Init(*configFile)
	if err != nil {
		log.Panic(err)
	}

	db, err := gorm.Open(sqlite.Open(cfg.DBDsn), &gorm.Config{})
	if err != nil {
		log.Panic("failed to connect database")
	}

	if err = model.Init(db); err != nil {
		log.Panic(err)
	}
	r := gin.Default()
	ctx := &core.GHPRContext{
		Gin: r,
		Db:  db,
		Cfg: cfg,
	}

	err = route.Init(ctx)
	if err != nil {
		log.Panic(err)
	}
	err = r.Run(cfg.ListenAddr)
	if err != nil {
		log.Panic(err)
	}
}
