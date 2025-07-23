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
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"github.com/1995parham-teaching/tinyurl/internal/domain/model/url"
	"github.com/1995parham-teaching/tinyurl/internal/domain/service/urlsvc"
	"github.com/1995parham-teaching/tinyurl/internal/infra/http/handler"
)

type MockURLSvc struct {
	mock.Mock
}

func (m *MockURLSvc) Create(ctx context.Context, longURL string, expire *time.Time) (string, error) {
	args := m.Called(ctx, longURL, expire)

	return args.String(0), args.Error(1)
}

func (m *MockURLSvc) CreateWithKey(ctx context.Context, key, longURL string, expire *time.Time) error {
	args := m.Called(ctx, key, longURL, expire)

	return args.Error(0)
}

func (m *MockURLSvc) Visit(ctx context.Context, key string) (url.URL, error) {
	args := m.Called(ctx, key)

	return args.Get(0).(url.URL), args.Error(1) // nolint: forcetypeassert
}

// nolint: funlen
func TestURL_Create(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		body               string
		mockSvc            func(*MockURLSvc)
		expectedStatusCode int
		expectedBody       string
		hasError           bool
	}{
		{
			name: "success with random key",
			body: `{"url": "http://example.com"}`,
			mockSvc: func(m *MockURLSvc) {
				m.On("Create", mock.Anything, "http://example.com", mock.AnythingOfType("*time.Time")).Return("random-key", nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `"random-key"`,
			hasError:           false,
		},
		{
			name: "success with custom key",
			body: `{"url": "http://example.com", "name": "custom-key"}`,
			mockSvc: func(m *MockURLSvc) {
				m.On("CreateWithKey",
					mock.Anything,
					"custom-key",
					"http://example.com",
					mock.AnythingOfType("*time.Time"),
				).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
			hasError:           false,
			expectedBody:       "",
		},
		{
			name: "duplicate key",
			body: `{"url": "http://example.com", "name": "duplicate-key"}`,
			mockSvc: func(m *MockURLSvc) {
				m.On("CreateWithKey",
					mock.Anything,
					"duplicate-key",
					"http://example.com",
					mock.AnythingOfType("*time.Time"),
				).Return(urlsvc.ErrKeyAlreadyExists)
			},
			expectedStatusCode: http.StatusBadRequest,
			hasError:           true,
			expectedBody:       "",
		},
		{
			name: "internal error",
			body: `{"url": "http://example.com", "name": "internal-error"}`,
			mockSvc: func(m *MockURLSvc) {
				m.On("CreateWithKey",
					mock.Anything,
					"internal-error",
					"http://example.com",
					mock.AnythingOfType("*time.Time"),
				).Return(errors.New("internal error")) // nolint: err113
			},
			expectedStatusCode: http.StatusInternalServerError,
			hasError:           true,
			expectedBody:       "",
		},
		{
			name:               "invalid request body",
			body:               `{"url": ""}`,
			mockSvc:            func(m *MockURLSvc) {},
			expectedStatusCode: http.StatusBadRequest,
			hasError:           true,
			expectedBody:       "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/urls", strings.NewReader(tc.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockSvc := new(MockURLSvc)
			tc.mockSvc(mockSvc)

			h := handler.URL{
				Service: mockSvc,
				Logger:  zap.NewNop(),
				Tracer:  noop.NewTracerProvider().Tracer(""),
			}

			err := h.Create(c)
			if tc.hasError {
				require.Error(err)

				var he *echo.HTTPError

				errors.As(err, &he)
				require.Equal(tc.expectedStatusCode, he.Code)
			} else {
				require.NoError(err)
				require.Equal(tc.expectedStatusCode, rec.Code)
			}

			if tc.expectedBody != "" {
				var expected, actual any

				err = json.Unmarshal([]byte(tc.expectedBody), &expected)
				require.NoError(err)
				err = json.Unmarshal(rec.Body.Bytes(), &actual)
				require.NoError(err)
				require.Equal(expected, actual)
			}

			mockSvc.AssertExpectations(t)
		})
	}
}

// nolint: funlen
func TestURL_Retrieve(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		key                string
		mockSvc            func(*MockURLSvc)
		expectedStatusCode int
		expectedLocation   string
		hasError           bool
	}{
		{
			name: "success",
			key:  "test-key",
			mockSvc: func(m *MockURLSvc) {
				m.On("Visit", mock.Anything, "test-key").Return(url.URL{URL: "http://example.com"}, nil) // nolint: exhaustruct
			},
			expectedStatusCode: http.StatusFound,
			expectedLocation:   "http://example.com",
			hasError:           false,
		},
		{
			name: "not found",
			key:  "not-found-key",
			mockSvc: func(m *MockURLSvc) {
				m.On("Visit", mock.Anything, "not-found-key").Return(url.URL{}, urlsvc.ErrURLNotFound) // nolint: exhaustruct
			},
			expectedStatusCode: http.StatusNotFound,
			expectedLocation:   "",
			hasError:           true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require := require.New(t)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/"+tc.key, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("key")
			c.SetParamValues(tc.key)

			mockSvc := new(MockURLSvc)
			tc.mockSvc(mockSvc)

			h := handler.URL{
				Service: mockSvc,
				Logger:  zap.NewNop(),
				Tracer:  noop.NewTracerProvider().Tracer(""),
			}

			err := h.Retrieve(c)
			if tc.hasError {
				require.Error(err)

				var he *echo.HTTPError

				errors.As(err, &he)
				require.Equal(tc.expectedStatusCode, he.Code)
			} else {
				require.NoError(err)
				require.Equal(tc.expectedStatusCode, rec.Code)
			}

			if tc.expectedLocation != "" {
				assert.Equal(t, tc.expectedLocation, rec.Header().Get(echo.HeaderLocation))
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
