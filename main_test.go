package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"time"

	"testing"
)

type MockedRedisClient struct {
	mock.Mock
}

func (m *MockedRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return redis.NewStringResult(args.String(0), args.Error(1))
}

func (m *MockedRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return redis.NewStatusCmd(ctx, args.String(0))
}

func Test_redirectToHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()

	// Create a mock Redis client
	mockRedis := new(MockedRedisClient)

	mockRedis.On("Get", mock.Anything, "test-path").Return(redis.NewStringResult("https://example.com", nil))

	// Create a new App instance with the mock Redis client
	//redisService := &RedisClient{client: mockRedis}
	//
	//router.GET("/redirectTo/:path", func(c *gin.Context) {
	//	err := redisService.redirectToHandler(c)
	//	assert.NoError(t, err)
	//})

	req, _ := http.NewRequest("GET", "/redirectTo/test-path", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status code %d, but got %d", http.StatusFound, w.Code)
	}

	expectedLocation := "https://example.com"
	if w.Header().Get("Location") != expectedLocation {
		t.Errorf("Expected redirect to %s, but got %s", expectedLocation, w.Header().Get("Location"))
	}

	// Assert that the mock was called
	mockRedis.AssertCalled(t, "Get", mock.Anything, "test-path")

}

func Test_appendHttpsToUrl(t *testing.T) {
	type args struct {
		foundKey string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no prefix should add HTTPS",
			args: args{foundKey: "example.com"},
			want: "https://example.com",
		},
		{
			name: "HTTPS should do nothing",
			args: args{foundKey: "https://example.com"},
			want: "https://example.com",
		},
		{
			name: "HTTP should do nothing",
			args: args{foundKey: "http://example.com"},
			want: "http://example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appendHttpsToUrl(tt.args.foundKey); got != tt.want {
				t.Errorf("appendHttpsToUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
