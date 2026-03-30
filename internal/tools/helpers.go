package tools

import (
	"time"

	osc "github.com/outscale/osc-sdk-go/v2"
)

// safeString safely extracts a string from a pointer.
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// safeInt safely extracts an int from a pointer.
func safeInt(i *int32) int {
	if i == nil {
		return 0
	}
	return int(*i)
}

// safeInt32 safely extracts an int32 from a pointer.
func safeInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func safeBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// safeInt64 safely extracts an int64 from a pointer.
func safeInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// safeTime safely extracts a time string from a pointer.
func safeTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// safeResponseId extracts the request ID from a ResponseContext.
func safeResponseId(ctx *osc.ResponseContext) string {
	if ctx == nil || ctx.RequestId == nil {
		return ""
	}
	return *ctx.RequestId
}
