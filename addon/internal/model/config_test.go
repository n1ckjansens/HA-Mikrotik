package model

import "testing"

func TestRouterConfigBaseURL(t *testing.T) {
	t.Helper()

	tests := []struct {
		name string
		cfg  RouterConfig
		want string
	}{
		{
			name: "plain host with ssl disabled",
			cfg:  RouterConfig{Host: "192.168.88.1", SSL: false},
			want: "http://192.168.88.1/rest",
		},
		{
			name: "plain host with ssl enabled",
			cfg:  RouterConfig{Host: "192.168.88.1", SSL: true},
			want: "https://192.168.88.1/rest",
		},
		{
			name: "host with explicit scheme keeps scheme",
			cfg:  RouterConfig{Host: "http://192.168.88.1", SSL: true},
			want: "http://192.168.88.1/rest",
		},
		{
			name: "host with rest path does not duplicate rest",
			cfg:  RouterConfig{Host: "192.168.88.1/rest", SSL: true},
			want: "https://192.168.88.1/rest",
		},
		{
			name: "host with explicit rest path does not duplicate rest",
			cfg:  RouterConfig{Host: "https://192.168.88.1/rest", SSL: true},
			want: "https://192.168.88.1/rest",
		},
		{
			name: "host with custom path appends rest",
			cfg:  RouterConfig{Host: "https://192.168.88.1/api", SSL: true},
			want: "https://192.168.88.1/api/rest",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			got := tt.cfg.BaseURL()
			if got != tt.want {
				t.Fatalf("BaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
