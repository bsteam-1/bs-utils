package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logs struct {
	logger       *zap.Logger
	sugar        *zap.SugaredLogger
	config       zap.Config
	path         string
	filename     string
	pullfilename string
	debug        bool
}

// 사용법 :
//
// log1 :=  &logs{
// path:     "./log/upload/",
// filename: "Success",
// }
// log1.Init()
func (l *Logs) Init() {
	os.MkdirAll(l.path, 0755)
	l.config = zap.NewProductionConfig()
	encoderconfigz := zapcore.EncoderConfig{
		TimeKey:        "date",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	l.config.EncoderConfig = encoderconfigz

	l.create()

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			camse := Join(l.filename, "_", dateformats())
			if l.pullfilename != camse {
				if l.debug {
					log.Println("로그_상태_새로운파일생성")
				}
				l.create()
			}
		}
	}()
	go func() {
		for {
			go func() {
				defer recover()
				findoldfile(l.path)
			}()
			time.Sleep(time.Hour)
		}
	}()
}
func (l *Logs) create() {
	filename := Join(l.path, l.filename, "_", dateformats())
	l.pullfilename = Join(l.filename, "_", dateformats())
	l.config.OutputPaths = []string{
		filename,
	}
	logger, err := l.config.Build()
	if err != nil {
		if l.debug {
			log.Println("로그_상태_생성실패", err)
		}
	}
	l.logger = logger
	l.sugar = l.logger.Sugar()
	err = l.logger.Sync()
	if err != nil {
		if l.debug {
			log.Println("로그_상태_오류 : ", err)
		}
	}
}
func findoldfile(filepathname string) {
	df := time.Now()
	filepath.Walk(filepathname, func(pathi string, infoi os.FileInfo, err error) error {
		reg := regexp.MustCompile(`-(\w+-\w+-\w+).log`)
		s := reg.FindStringSubmatch(pathi)
		if len(s) > 0 {
			t, _ := time.ParseInLocation("2006-01-02", s[1], df.Location())
			if int(df.Sub(t).Hours()) > 168 {
				os.Remove(pathi)
			}
		}
		return err
	})
}

func dateformats() string {
	t := time.Now()
	datez := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
	return datez
}
