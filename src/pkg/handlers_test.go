package pkg

import "testing"

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
