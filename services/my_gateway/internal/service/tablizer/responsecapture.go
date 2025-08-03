package tablizer

import (
	"bytes"
	"net/http"
)

// ResponseCapture wraps http.ResponseWriter to capture response
type ResponseCapture struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	captured   bool
}

func NewResponseCapture(w http.ResponseWriter) *ResponseCapture {
	return &ResponseCapture{
		ResponseWriter: w,
		body:           new(bytes.Buffer),
		statusCode:     200,
	}
}

func (rc *ResponseCapture) Write(b []byte) (int, error) {
	if !rc.captured {
		return rc.body.Write(b)
	}
	return rc.ResponseWriter.Write(b)
}

func (rc *ResponseCapture) WriteHeader(statusCode int) {
	if !rc.captured {
		rc.statusCode = statusCode
		return
	}
	rc.ResponseWriter.WriteHeader(statusCode)
}

func (rc *ResponseCapture) GetBody() []byte {
	return rc.body.Bytes()
}

func (rc *ResponseCapture) GetStatusCode() int {
	return rc.statusCode
}

func (rc *ResponseCapture) FlushOriginal() {
	rc.captured = true
	rc.ResponseWriter.WriteHeader(rc.statusCode)
	rc.ResponseWriter.Write(rc.body.Bytes())
}

func (rc *ResponseCapture) FlushCSV(csvData []byte) {
	rc.captured = true
	rc.ResponseWriter.Header().Set("Content-Type", "text/csv")
	rc.ResponseWriter.Header().Set("Content-Disposition", "attachment; filename=export.csv")
	rc.ResponseWriter.WriteHeader(200)
	rc.ResponseWriter.Write(csvData)
}
