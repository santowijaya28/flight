package response

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteSuccessResponse(t *testing.T) {
	tests := []struct {
		name       string
		data       any
		statusCode int
		wantStatus int
		wantBody   string
	}{
		{
			name: "default status",
			data: struct {
				Message string `json:"message"`
			}{Message: "ok"},
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"ok"}`,
		},
		{
			name:       "custom status",
			data:       map[string]string{"message": "created"},
			statusCode: http.StatusCreated,
			wantStatus: http.StatusCreated,
			wantBody:   `{"message":"created"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			WriteSuccessResponse(recorder, Response{
				Data:           tt.data,
				HttpStatusCode: tt.statusCode,
			})

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("content type = %q, want application/json", got)
			}
			if recorder.Body.String() != tt.wantBody {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestWriteErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		err        string
		statusCode int
		wantStatus int
		wantBody   string
	}{
		{
			name:       "bad request",
			err:        "invalid input",
			statusCode: http.StatusBadRequest,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"header":{"error":"invalid input"}}`,
		},
		{
			name:       "server error",
			err:        "all provider down",
			statusCode: http.StatusInternalServerError,
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"header":{"error":"all provider down"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			WriteErrorResponse(recorder, Response{
				Error:          tt.err,
				HttpStatusCode: tt.statusCode,
			})

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("content type = %q, want application/json", got)
			}
			if recorder.Body.String() != tt.wantBody {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), tt.wantBody)
			}
		})
	}
}
