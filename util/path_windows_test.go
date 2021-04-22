package util

import (
	"os"
	"testing"
)

func TestGetPath(t *testing.T) {
	_ = os.Setenv("BASE", `C:\abc`)
	os.Setenv("USER", "normal")

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			"plain",
			`C:\User\user`,
			`C:\User\user`,
		},
		{
			"one env",
			`%BASE%\user`,
			`C:\abc\user`,
		},
		{
			"two env",
			`%BASE%\%USER%`,
			`C:\abc\normal`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPath(tt.arg); got != tt.want {
				t.Errorf("GetPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
