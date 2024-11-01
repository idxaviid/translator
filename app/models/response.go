package models

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// MyJSONResponse respuesta JSON standard
type MyJSONResponse struct {
	MyResponse
	status      int
	contentType string
	writer      http.ResponseWriter
}

// MyResponse resposta base
type MyResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"body"`
	Msg  string      `json:"msg"`
}

// CreateJSONResponse crea una respuesta a partir de r MyResponse en w
func CreateJSONResponse(w http.ResponseWriter, r MyResponse, statusCode int) MyJSONResponse {
	jsonResponse := MyJSONResponse{status: statusCode, contentType: "application/json", writer: w}
	jsonResponse.MyResponse.Data = r.Data
	jsonResponse.MyResponse.Code = r.Code
	jsonResponse.MyResponse.Msg = r.Msg
	return jsonResponse
}

// SendJSONResponse escribe la respuesta JSON en el writer
func (my *MyJSONResponse) SendJSONResponse() {
	my.writer.Header().Set("Content-Type", my.contentType)
	my.writer.WriteHeader(my.status)

	output, _ := json.Marshal(&my)
	fmt.Fprintln(my.writer, string(output))
}

// GenerateResponse GenerateResponse
func GenerateResponse(w http.ResponseWriter, r MyResponse, status int) {
	MyJSONResponse := CreateJSONResponse(w, r, status)
	MyJSONResponse.SendJSONResponse()
}

// MyResponseWriter custom response with minimal propoerties
type MyResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// NewResponseWriter NewResponseWriter
func NewResponseWriter(w http.ResponseWriter) *MyResponseWriter {
	return &MyResponseWriter{w, http.StatusOK}
}

// MyHTMLResponse respuesta html
type MyHTMLResponse struct {
	content string
	status  int
	writer  http.ResponseWriter
}

// GenerateHTMLResponse escribe la respuesta en el writer
func GenerateHTMLResponse(w http.ResponseWriter, content string, statusCode int) {
	response := MyHTMLResponse{status: statusCode, writer: w}
	response.writer.Header().Set("Content-Type", "text/html;charset=ISO-8859-1")
	response.writer.WriteHeader(response.status)
	fmt.Fprint(response.writer, content)
}
