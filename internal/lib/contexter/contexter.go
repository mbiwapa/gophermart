package contexter

import "context"

type ctxKey string

var (
	// RequestId is the context key for the request ID.
	RequestId = ctxKey("request_id")
)

// GetRequestId	returns the request id from the context.
func GetRequestId(ctx context.Context) string {
	requestId := ctx.Value(RequestId).(string)
	return requestId
}
