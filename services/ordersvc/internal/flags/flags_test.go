package flags_test

import (
	"context"
	"testing"

	"rxw1/ordersvc/internal/flags"
)

func TestFlags_RedisEnabled(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.New()
			got := f.ThrottleEnabled(context.Background())
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("ThrottleEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
