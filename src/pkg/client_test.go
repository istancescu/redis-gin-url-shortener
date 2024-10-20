package pkg

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppendHttpsToUrl(t *testing.T) {
	type args struct {
		foundKey string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no prefixHttp should add HTTPS",
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

func TestRedirectToHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock Redis client
	mockClient, mockRedis := redismock.NewClientMock()

	client := &RedisClient{
		Client: mockClient,
	}

	type args struct {
		foundKey string
	}
	tests := []struct {
		name               string
		args               args
		want               string
		expectedStatusCode int
		mockRedisAction    func()
	}{
		{
			name:               "no prefix Http should add HTTPS",
			args:               args{foundKey: "1234"},
			want:               "https://example.com",
			expectedStatusCode: http.StatusFound,
			mockRedisAction:    func() { mockRedis.ExpectGet("1234").SetVal("example.com") },
		},
		{
			name:               "HTTP should still give HTTP",
			args:               args{foundKey: "1234"},
			want:               "http://example.com",
			expectedStatusCode: http.StatusFound,
			mockRedisAction:    func() { mockRedis.ExpectGet("1234").SetVal("http://example.com") },
		},
		{
			name:               "HTTP should do nothing",
			args:               args{foundKey: "1234"},
			want:               "https://example.com",
			expectedStatusCode: http.StatusFound,
			mockRedisAction:    func() { mockRedis.ExpectGet("1234").SetVal("https://example.com") },
		},
		{
			name:               "should break",
			args:               args{foundKey: "1234"},
			want:               "",
			expectedStatusCode: http.StatusInternalServerError,
			mockRedisAction:    func() { mockRedis.ExpectGet("1234").SetErr(redis.Nil) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()

			tt.mockRedisAction()

			req, _ := http.NewRequest("GET", "/redirectTo/"+tt.args.foundKey, nil)

			router.GET("/redirectTo/:path", RedirectToHandler(client))

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			expectedLocation := tt.want

			assert.NoError(t, mockRedis.ExpectationsWereMet())
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, expectedLocation, w.Header().Get("Location"))
			mockRedis.ClearExpect()

		})
	}

}

func TestRedisClient_DefaultPathHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a mock Redis client
	mockClient, mockRedis := redismock.NewClientMock()

	client := &RedisClient{
		Client: mockClient,
	}

	type args struct {
		foundKeyVal string
	}
	tests := []struct {
		name               string
		args               args
		want               string
		expectedStatusCode int
		mockRedisAction    func()
	}{
		{
			name:               "Redis found key should return 200",
			want:               "example.com",
			args:               args{foundKeyVal: "example.com"},
			expectedStatusCode: http.StatusOK,
			mockRedisAction:    func() { mockRedis.ExpectGet("example.com").SetVal("1234") },
		},
		{
			name:               "Redis not found key should create (201)",
			want:               "example.com",
			args:               args{foundKeyVal: "example.com"},
			expectedStatusCode: http.StatusCreated,
			mockRedisAction:    func() { mockRedis.ExpectGet("example.com").SetErr(redis.Nil) },
		},
		{
			name:               "Redis broke with proto/other error",
			args:               args{foundKeyVal: "1234"},
			want:               "example.com",
			expectedStatusCode: http.StatusInternalServerError,
			mockRedisAction:    func() { mockRedis.ExpectGet("example.com").SetErr(redis.TxFailedErr) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()

			tt.mockRedisAction()

			req, _ := http.NewRequest("GET", "/url/"+tt.args.foundKeyVal, nil)

			router.GET("/url/:urlToShorten", DefaultPathHandler(client))

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			mockRedis.ClearExpect()

		})
	}

}
