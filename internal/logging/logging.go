// Package logging configures the process-wide structured (JSON) slog logger
// used by the Lambda handler path.
//
// CloudWatch Logs Insights auto-discovers top-level keys of JSON log events as
// queryable fields, so emitting JSON here makes `message`, `path`, `requestId`
// etc. filterable directly (unlike plain fmt.Printf text, where only the
// system `@message` field exists). The Go runtime is provided.al2023, so
// Lambda's own JSON log format does not restructure application stdout — we
// must emit JSON ourselves.
package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// lambdaHandler decorates every record with the Lambda invocation's
// AwsRequestID (when present in ctx) so each log line correlates to the
// START/END/REPORT block for that invocation.
type lambdaHandler struct {
	slog.Handler
}

func (h lambdaHandler) Handle(ctx context.Context, r slog.Record) error {
	if lc, ok := lambdacontext.FromContext(ctx); ok && lc.AwsRequestID != "" {
		r.AddAttrs(slog.String("requestId", lc.AwsRequestID))
	}
	return h.Handler.Handle(ctx, r)
}

func (h lambdaHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return lambdaHandler{h.Handler.WithAttrs(attrs)}
}

func (h lambdaHandler) WithGroup(name string) slog.Handler {
	return lambdaHandler{h.Handler.WithGroup(name)}
}

// Init installs a JSON slog logger (writing to stdout) as the default,
// wrapping records with the Lambda request ID. Call once at process start.
func Init() {
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(lambdaHandler{Handler: base}))
}
