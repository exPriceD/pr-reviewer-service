package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"

	loggermocks "github.com/exPriceD/pr-reviewer-service/internal/domain/logger/mocks"
)

func TestLimitBodySize(t *testing.T) {
	tests := []struct {
		name        string
		maxBodySize int64
		bodySize    int
		wantStatus  int
	}{
		{
			name:        "small body",
			maxBodySize: 1000,
			bodySize:    100,
			wantStatus:  http.StatusOK,
		},
		{
			name:        "exact size",
			maxBodySize: 100,
			bodySize:    100,
			wantStatus:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := LimitBodySize(tt.maxBodySize)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body := make([]byte, tt.bodySize)
				_, _ = r.Body.Read(body)
				w.WriteHeader(http.StatusOK)
			}))

			body := make([]byte, tt.bodySize)
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			req.ContentLength = int64(tt.bodySize)

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestRecovery(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus int
		wantPanic  bool
	}{
		{
			name: "no panic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantStatus: http.StatusOK,
			wantPanic:  false,
		},
		{
			name: "panic recovered",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("test panic")
			},
			wantStatus: http.StatusInternalServerError,
			wantPanic:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := loggermocks.NewMockLogger(ctrl)
			if tt.wantPanic {
				ctxLogger := loggermocks.NewMockLogger(ctrl)
				logger.EXPECT().WithContext(gomock.Any()).Return(ctxLogger)
				ctxLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
			}

			handler := Recovery(logger)(tt.handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantPanic {
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", contentType)
				}
			}
		})
	}
}

func TestLogger(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		method     string
		path       string
	}{
		{
			name:       "success request",
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			path:       "/test",
		},
		{
			name:       "error request",
			statusCode: http.StatusBadRequest,
			method:     http.MethodPost,
			path:       "/test",
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			method:     http.MethodGet,
			path:       "/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := loggermocks.NewMockLogger(ctrl)
			ctxLogger := loggermocks.NewMockLogger(ctrl)
			logger.EXPECT().WithContext(gomock.Any()).Return(ctxLogger)

			switch {
			case tt.statusCode >= 500:
				ctxLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
			case tt.statusCode >= 400:
				ctxLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
			default:
				ctxLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			}

			handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestRequestID(t *testing.T) {
	tests := []struct {
		name          string
		headerValue   string
		chiRequestID  string
		wantHeaderSet bool
	}{
		{
			name:          "no request ID",
			headerValue:   "",
			chiRequestID:  "",
			wantHeaderSet: false,
		},
		{
			name:          "request ID from header",
			headerValue:   "test-request-id",
			chiRequestID:  "",
			wantHeaderSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.headerValue != "" {
				req.Header.Set("X-Request-ID", tt.headerValue)
			}

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			headerValue := w.Header().Get("X-Request-ID")
			if tt.wantHeaderSet {
				if headerValue == "" {
					t.Error("expected X-Request-ID header to be set")
				}
			}
		})
	}
}
