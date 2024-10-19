package config

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestCreateNewRedisConfig(t *testing.T) {
	type args struct {
		path        string
		shouldError bool
		errorText   string
	}
	tests := []struct {
		name string
		args args
		want *redis.Options
	}{
		{
			name: "Should successfully read a file",
			args: args{path: "../../config.yaml"},
			want: &redis.Options{
				Addr:     "127.0.0.1:6379",
				Password: "1234",
				DB:       0,
			},
		},
		{
			name: "Should throw path error",
			args: args{path: "montana", shouldError: true, errorText: "Couldn't find file for path: montana \n"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.args.shouldError {
				if got, _ := ProvideRedisConfig(tt.args.path); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ProvideRedisConfig() = %v, want %v", got, tt.want)
				}
			} else {
				assert.Error(t, fmt.Errorf(""), func() {
					_, _ = ProvideRedisConfig(tt.args.path)
				})
			}

		})
	}
}
