package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"github.com/1995parham-teaching/tinyurl/internal/domain/model/url"
	"github.com/1995parham-teaching/tinyurl/internal/domain/service/urlsvc"
	"github.com/1995parham-teaching/tinyurl/internal/infra/http/handler"
)

type MockURLSvc struct {
	mock.Mock
}

func (m *MockURLSvc) Create(ctx context.Context, longURL string, expire time.Duration) (string, error) {
	args := m.Called(ctx, longURL, expire)
	return args.String(0), args.Error(1)
}

func (m *MockURLSvc) CreateWithKey(ctx context.Context, key, longURL string, expire time.Duration) error {
	args := m.Called(ctx, key, longURL, expire)
	return args.Error(0)
}

func (m *MockURLSvc) Visit(ctx context.Context, key string) (url.URL, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(url.URL), args.Error(1)
}

func TestURL_Create(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		body               string
		mockSvc            func(*MockURLSvc)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "success with random key",
			body: `{"url": "http://example.com"}`,
			mockSvc: func(m *MockURLSvc) {
				m.On("Create", mock.Anything, "http://example.com", time.Duration(0)).Return("random-key", nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `"random-key"`,
		},
		{
			name: "success with custom key",
			body: `{"url": "http://example.com", "name": "custom-key"}`,
			mockSvc: func(m *MockURLSvc) {
				m.On("CreateWithKey", mock.Anything, "custom-key", "http://example.com", time.Duration(0)).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name: "duplicate key",
			body: `{"url": "http://example.com", "name": "duplicate-key"}`,
			mockSvc: func(m *MockURLSvc) {
				m.On("CreateWithKey", mock.Anything, "duplicate-key", "http://example.com", time.Duration(0)).Return(urlsvc.ErrKeyAlreadyExists)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "invalid request body",
			body:               `{"url": ""}`,
			mockSvc:            func(m *MockURLSvc) {},
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/urls", strings.NewReader(tc.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockSvc := new(MockURLSvc)
			tc.mockSvc(mockSvc)

			h := handler.URL{
				Store:  mockSvc,
				Logger: zap.NewNop(),
				Tracer: noop.NewTracerProvider().Tracer("test"),
			}

			err := h.Create(c)

			if assert.NoError(t, err) {
				assert.Equal(t, tc.expectedStatusCode, rec.Code)
				if tc.expectedBody != "" {
					var expected, actual interface{}
					err = json.Unmarshal([]byte(tc.expectedBody), &expected)
					assert.NoError(t, err)
					err = json.Unmarshal(rec.Body.Bytes(), &actual)
					assert.NoError(t, err)
					assert.Equal(t, expected, actual)
				}
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

func TestURL_Retrieve(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		key                string
		mockSvc            func(*MockURLSvc)
		expectedStatusCode int
		expectedLocation   string
	}{
		{
			name: "success",
			key:  "test-key",
			mockSvc: func(m *MockURLSvc) {
				m.On("Visit", mock.Anything, "test-key").Return(urlsvc.URL{URL: "http://example.com"}, nil)
			},
			expectedStatusCode: http.StatusFound,
			expectedLocation:   "http://example.com",
		},
		{
			name: "not found",
			key:  "not-found-key",
			mockSvc: func(m *MockURLSvc) {
				m.On("Visit", mock.Anything, "not-found-key").Return(urlsvc.URL{}, errors.New("not found"))
			},
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/"+tc.key, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("key")
			c.SetParamValues(tc.key)

			mockSvc := new(MockURLSvc)
			tc.mockSvc(mockSvc)

			h := handler.URL{
				Store:  mockSvc,
				Logger: zap.NewNop(),
				Tracer: trace.NewNoopTracerProvider().Tracer("test"),
			}

			err := h.Retrieve(c)

			if assert.NoError(t, err) {
				assert.Equal(t, tc.expectedStatusCode, rec.Code)
				if tc.expectedLocation != "" {
					assert.Equal(t, tc.expectedLocation, rec.Header().Get(echo.HeaderLocation))
				}
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
