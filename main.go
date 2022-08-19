package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wI2L/jsondiff"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func diffHandler(writer http.ResponseWriter, request *http.Request) {

	var now = time.Now()
	fmt.Printf("Comparing Files [%+v]\n", now)
	err := request.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println(err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	formdata := request.MultipartForm
	file1 := formdata.File["file1"]

	file1Json, err := file1[0].Open()
	defer file1Json.Close()
	if err != nil {
		log.Println(err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	file2 := formdata.File["file2"]
	file2Json, err := file2[0].Open()
	defer file2Json.Close()
	if err != nil {
		log.Println(err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	file1JsonByteValue, err := io.ReadAll(file1Json)

	if err != nil {
		log.Println(err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	file2JsonByteValue, err := io.ReadAll(file2Json)

	if err != nil {
		log.Println(err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	var json1 map[string]interface{}
	json.Unmarshal(file1JsonByteValue, &json1)
	var json2 map[string]interface{}
	json.Unmarshal(file2JsonByteValue, &json2)
	patch, err := jsondiff.Compare(json1, json2)

	if err != nil {
		log.Println(err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	var result bytes.Buffer
	result.WriteString("[")
	for _, op := range patch {
		res := fmt.Sprintf("%s,\n", op)
		fmt.Printf(res)
		result.WriteString(res)

	}
	result.WriteString("{}")
	result.WriteString("]")

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusAccepted)

	writer.Write(result.Bytes())

}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func main() {

	appPort := ":" + getenv("PORT", "3005")
	http.HandleFunc("/api/diff", diffHandler)
	http.ListenAndServe(appPort, nil)
}
