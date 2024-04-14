package apiresponse

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type APIError struct {
	Message string `json:"message"`
}

func StatusText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 500:
		return "Internal Server Error"
	default:
		return "Unknown"
	}
}

func OK200(w http.ResponseWriter, data any) {
	writeJSONResponse(w, 200, data)
}

func NotFound404(w http.ResponseWriter) {
	writeJSONResponse(w, 404, StatusText(404))
}

func InternalError500(w http.ResponseWriter) {
	writeJSONResponse(w, 500, StatusText(500))
}

func Error(w http.ResponseWriter, message string, code int) {
	writeJSONResponse(w, code, APIError{Message: message})
}

func writeJSONResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	switch {
	case code < 200:
		panic(fmt.Sprintf("code %d must be >= 200", code))
	case data == nil:
		data = &struct {
			Message string `json:"message"`
		}{Message: StatusText(code)}
	}

	b, err := json.Marshal(data)
	if err != nil {
		fields := log.Fields{"data": data, "code": code}
		log.WithFields(fields).Errorf("%+v", err)
		w.WriteHeader(500)
		w.Write([]byte(`{"message":"Internal Server Error"}`))
		return
	}

	w.WriteHeader(code)
	w.Write(b)
}
