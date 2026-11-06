package aistudio

import (
	"errors"
	"net/http"
	"testing"
)

func TestAssistantResumeRetryable(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "sensitive words 502",
			err:  &RPCError{Method: assistantStreamMethod, StatusCode: http.StatusBadGateway, Message: "生成步骤失败: Sensitive words detected"},
			want: true,
		},
		{
			name: "other 502",
			err:  &RPCError{Method: assistantStreamMethod, StatusCode: http.StatusBadGateway, Message: "Internal server error"},
		},
		{
			name: "sensitive words other status",
			err:  &RPCError{Method: assistantStreamMethod, StatusCode: http.StatusBadRequest, Message: "Sensitive words detected"},
		},
		{
			name: "sensitive words other method",
			err:  &RPCError{Method: assistantOfflineMethod, StatusCode: http.StatusBadGateway, Message: "Sensitive words detected"},
		},
		{
			name: "interrupted stream",
			err:  &StreamInterruptedError{Method: assistantStreamMethod, Err: errors.New("connection reset")},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := assistantResumeRetryable(test.err); got != test.want {
				t.Fatalf("assistantResumeRetryable() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestAssistantResumeAttempts(t *testing.T) {
	for _, test := range []struct {
		name  string
		env   map[string]string
		want  int
	}{
		{name: "disabled", want: 0},
		{name: "enabled default", env: map[string]string{"MAG_BUILD_AUTO_CONTINUE": "true"}, want: 3},
		{name: "configured", env: map[string]string{"MAG_BUILD_AUTO_CONTINUE": "TRUE", "MAG_BUILD_AUTO_CONTINUE_MAX": "5"}, want: 5},
		{name: "invalid max", env: map[string]string{"MAG_BUILD_AUTO_CONTINUE": "true", "MAG_BUILD_AUTO_CONTINUE_MAX": "bad"}, want: 0},
		{name: "capped", env: map[string]string{"MAG_BUILD_AUTO_CONTINUE": "true", "MAG_BUILD_AUTO_CONTINUE_MAX": "99"}, want: 10},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("MAG_BUILD_AUTO_CONTINUE", "")
			t.Setenv("MAG_BUILD_AUTO_CONTINUE_MAX", "")
			for key, value := range test.env {
				t.Setenv(key, value)
			}
			if got := assistantResumeAttempts(); got != test.want {
				t.Fatalf("assistantResumeAttempts() = %d, want %d", got, test.want)
			}
		})
	}
}
