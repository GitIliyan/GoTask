package internal

import (
	"encoding/json"
	"net/http"
)

type ErrorFrequency struct {
	Expression string `json:"expression"`
	Endpoint   string `json:"endpoint"`
	Frequency  int    `json:"frequency"`
	Type       string `json:"type"`
}

func InitializeMap() {
	errorFrequencyMap = make(map[string]ErrorFrequency)
}

var errorFrequencyMap map[string]ErrorFrequency

func ExtractHandler(w http.ResponseWriter, req *http.Request) {
	handleRequest(w, req, false)
}

func ValidateHandler(w http.ResponseWriter, req *http.Request) {
	handleRequest(w, req, true)
}

func ErrorHandler(w http.ResponseWriter, err error, endpoint, expression string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	IncrementErrorFrequency(err.Error(), endpoint, expression)

	errorResponse := ErrorFrequency{
		Expression: expression,
		Endpoint:   endpoint,
		Frequency:  errorFrequencyMap[err.Error()].Frequency,
		Type:       err.Error(),
	}

	errorJSON, _ := json.Marshal(errorResponse)

	_, _ = w.Write(errorJSON)
}

func GetErrorFrequenciesHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	errorFrequencies := getErrorFrequencies()

	errorJSON, err := json.Marshal(errorFrequencies)
	if err != nil {
		ErrorHandler(w, err, "N/A", "N/A")
		return
	}

	_, _ = w.Write(errorJSON)
}
