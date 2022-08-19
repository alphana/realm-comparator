package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/wI2L/jsondiff"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

func diffHandler(writer http.ResponseWriter, request *http.Request) {

	var now = time.Now()
	logger.Debug("Comparing Files started at ", zap.Time("now", now))

	err := request.ParseMultipartForm(10 << 20)
	if err != nil {
		logger.Error("Couldn't Parse Request", zap.Error(err))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	formdata := request.MultipartForm
	file1 := formdata.File["file1"]

	file1Json, err := file1[0].Open()
	defer func(file1Json multipart.File) {
		err := file1Json.Close()
		if err != nil {
			logger.Error("Couldn't close file1", zap.Error(err))
		}
	}(file1Json)
	if err != nil {
		logger.Error("Couldn't read file1", zap.Error(err))

		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	file2 := formdata.File["file2"]
	file2Json, err := file2[0].Open()
	defer func(file2Json multipart.File) {
		err := file2Json.Close()
		if err != nil {
			logger.Error("Couldn't close file2", zap.Error(err))
		}
	}(file2Json)
	if err != nil {

		logger.Error("Couldn't read file2", zap.Error(err))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	file1JsonByteValue, err := io.ReadAll(file1Json)

	if err != nil {
		logger.Error("Couldn't parse file1 json", zap.Error(err))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	file2JsonByteValue, err := io.ReadAll(file2Json)

	if err != nil {
		logger.Error("Couldn't parse file2 json", zap.Error(err))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	var json1 map[string]interface{}
	err = json.Unmarshal(file1JsonByteValue, &json1)
	if err != nil {
		logger.Error("Couldn't parse file1", zap.Error(err))
		return
	}
	var json2 map[string]interface{}
	err = json.Unmarshal(file2JsonByteValue, &json2)
	if err != nil {
		logger.Error("Couldn't parse file2", zap.Error(err))
		return
	}

	patch, err := jsondiff.Compare(json1, json2)

	if err != nil {
		logger.Error("Couldn't compare", zap.Error(err))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	var result bytes.Buffer
	result.WriteString("[")
	for _, op := range patch {
		res := fmt.Sprintf("%s,", op)
		fmt.Printf(res)
		logger.Debug("diff ", zap.String("", res))
		result.WriteString(res)

	}
	result.WriteString("{}")
	result.WriteString("]")

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusAccepted)

	_, err = writer.Write(result.Bytes())
	if err != nil {
		logger.Error("Couldn't read file2", zap.Error(err))
		return
	}

	logger.Debug("Comparing Files took at ", zap.Duration("now", time.Now().Sub(now)))
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func main() {
	server()
}

var (
	// StdoutEncoderConfig is default zap logger encoder config whose output path is stdout.
	StdoutEncoderConfig = &zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	StdoutLoggerConfig = &zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:       true,
		Encoding:          "console",
		DisableStacktrace: true,
		EncoderConfig:     *StdoutEncoderConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	logger, _ = StdoutLoggerConfig.Build()
)

func server() {

	appPort := ":" + getenv("PORT", "3005")

	router := mux.NewRouter()
	router.HandleFunc("/api/diff", diffHandler)

	srv := &http.Server{
		Handler: router,
		Addr:    appPort,
	}
	logger.Info("Starting server at %s", zap.String("address", srv.Addr))
	log.Fatal(srv.ListenAndServe())
}
