package pkg

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
			if got := AppendHttpsToUrl(tt.args.foundKey); got != tt.want {
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
		foundKey    string
		foundKeyVal string
	}
	tests := []struct {
		name               string
		args               args
		want               string
		redisShouldBeNil   bool
		expectedStatusCode int
	}{
		{
			name:               "no prefix Http should add HTTPS",
			args:               args{foundKey: "1234", foundKeyVal: "example.com"},
			want:               "https://example.com",
			expectedStatusCode: http.StatusFound,
		},
		{
			name:               "HTTP should still give HTTP",
			args:               args{foundKey: "1234", foundKeyVal: "http://example.com"},
			want:               "http://example.com",
			expectedStatusCode: http.StatusFound,
		},
		{
			name:               "HTTP should do nothing",
			args:               args{foundKey: "1234", foundKeyVal: "https://example.com"},
			want:               "https://example.com",
			expectedStatusCode: http.StatusFound,
		},
		{
			name:               "should break",
			args:               args{foundKey: "1234", foundKeyVal: "https://example.com"},
			want:               "",
			expectedStatusCode: http.StatusInternalServerError,
			redisShouldBeNil:   true,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()

			if tt.redisShouldBeNil == true {
				mockRedis.ExpectGet(tt.args.foundKey).SetErr(redis.Nil)
			}

			if tt.redisShouldBeNil == false {
				mockRedis.ExpectGet(tt.args.foundKey).SetVal(tt.args.foundKeyVal)
			}

			req, _ := http.NewRequest("GET", "/redirectTo/"+tt.args.foundKey, nil)

			router.GET("/redirectTo/:path", client.RedirectToHandler)

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
		foundKey         string
		foundKeyVal      string
		redisShouldBeNil bool
	}
	tests := []struct {
		name               string
		args               args
		want               string
		expectedStatusCode int
	}{
		{
			name:               "1",
			args:               args{foundKey: "1234", foundKeyVal: "example.com", redisShouldBeNil: true},
			want:               "example.com",
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:               "2",
			args:               args{foundKey: "1234", foundKeyVal: "example.com", redisShouldBeNil: false},
			want:               "example.com",
			expectedStatusCode: http.StatusOK,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()

			if tt.args.redisShouldBeNil == true {
				mockRedis.ExpectGet(tt.args.foundKeyVal).SetErr(redis.Nil)
				mockRedis.ExpectSet(tt.args.foundKey, tt.args.foundKeyVal, time.Hour*25).SetVal("OK")
				mockRedis.ExpectSet(tt.args.foundKeyVal, tt.args.foundKey, time.Hour*25).SetVal("OK")
			} else {
				mockRedis.ExpectGet(tt.args.foundKeyVal).SetVal(tt.args.foundKey)
			}

			req, _ := http.NewRequest("GET", "/url/"+tt.args.foundKeyVal, nil)

			router.GET("/url/:urlToShorten", client.DefaultPathHandler)

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			//expectedLocation := tt.want

			//assert.NoError(t, mockRedis.ExpectationsWereMet())
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			//assert.Equal(t, expectedLocation, w.Header().Get("Location"))
			mockRedis.ClearExpect()

		})
	}

}
