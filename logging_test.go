/*
* Copyright 2022-2023 Thorsten A. Knieling
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*    http://www.apache.org/licenses/LICENSE-2.0
*
 */

package log

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func initTestLogWithFile(t *testing.T, fileName string) error {
	err := initLogLevelWithFile(fileName, zapcore.DebugLevel)
	if err != nil {
		t.Fatalf("error opening file: %v", err)
	}
	return err
}

func newWinFileSink(u *url.URL) (zap.Sink, error) {
	// Remove leading slash left by url.Parse()
	return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func initLogLevelWithFile(fileName string, level zapcore.Level) (err error) {
	p := os.Getenv("LOGPATH")
	if p == "" {
		p = os.TempDir()
	}
	var name string
	if runtime.GOOS == "windows" {
		zap.RegisterSink("winfile", newWinFileSink)
		//		OutputPaths: []string{"stdout", "winfile:///" + filepath.Join(GlobalConfigDir.Path, "info.log.json")},
		name = "winfile:///" + p + string(os.PathSeparator) + fileName
	} else {
		name = "file://" + filepath.ToSlash(p+string(os.PathSeparator)+fileName)
	}

	rawJSON := []byte(`{
	"level": "error",
	"encoding": "console",
	"outputPaths": [ "XXX"],
	"errorOutputPaths": ["stderr"],
	"encoderConfig": {
	  "messageKey": "message",
	  "levelKey": "level",
	  "levelEncoder": "lowercase"
	}
  }`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		return err
	}
	cfg.Level.SetLevel(level)
	cfg.OutputPaths = []string{name}
	logger, err := cfg.Build()
	if err != nil {
		return err
	}
	defer logger.Sync()

	sugar := logger.Sugar()
	Log = sugar

	sugar.Infof("AdabasGoApi logger initialization succeeded")
	return nil
}

func doTrackCall() {
	defer TimeTrack(time.Now(), "Time Track Unit test ")

}

func TestLogZap(t *testing.T) {
	fileName := "zap.log"
	os.Remove(os.TempDir() + "/" + fileName)
	err := initTestLogWithFile(t, fileName)
	if !assert.NoError(t, err) {
		fmt.Println(err)
		return
	}

	d := IsDebugLevel()
	SetDebugLevel(true)
	doTrackCall()

	hallo := "HELLO"
	Log.Debugf("This is a test of data %s", hallo)

	LogMultiLineString(true, "ABC\nXXXX\n")
	SetDebugLevel(false)
	SetDebugLevel(d)

	flog, err := os.Open(os.TempDir() + "/" + fileName)
	if !assert.NoError(t, err) {
		return
	}
	logInfo, err := io.ReadAll(flog)
	if !assert.NoError(t, err) {
		return
	}
	assert.Regexp(t, `info	AdabasGoApi logger initialization succeeded
info	Time Track Unit test  took \d*ns
debug	This is a test of data HELLO
debug	ABC
debug	XXXX
debug`, string(logInfo))
}

func TestLogrus(t *testing.T) {
	fileName := "logrus.log"
	os.Remove(os.TempDir() + "/" + fileName)
	log := logrus.StandardLogger()

	fmt.Println("Init logging")
	SetDebugLevel(true)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05",
	})
	log.SetLevel(logrus.DebugLevel)
	p := os.Getenv("LOGPATH")
	if p == "" {
		p = os.TempDir()
	}
	f, err := os.OpenFile(p+"/"+fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if !assert.NoError(t, err) {
		fmt.Println("Error opening log:", err)
		return
	}
	log.SetOutput(f)
	log.Infof("Init logrus")
	Log = log
	fmt.Println("Logging running")

	flog, err := os.Open(os.TempDir() + "/" + fileName)
	if !assert.NoError(t, err) {
		return
	}
	logInfo, err := io.ReadAll(flog)
	if !assert.NoError(t, err) {
		return
	}
	assert.Regexp(t, "time=\"20..-..-..T..:..:..\" level=info msg=\"Init logrus\"\n", string(logInfo))
}
