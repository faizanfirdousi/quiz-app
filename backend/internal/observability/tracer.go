package observability

import (
	"log/slog"

	"github.com/aws/aws-xray-sdk-go/xray"
)

// InitTracer initializes AWS X-Ray SDK.
// In local mode, if the X-Ray daemon is not running, it logs a warning and continues
// without tracing. In production, the Lambda execution environment provides the daemon.
func InitTracer(env string) {
	err := xray.Configure(xray.Config{
		LogLevel: "warn",
	})
	if err != nil {
		slog.Warn("X-Ray initialization failed — tracing disabled", "error", err.Error())
		return
	}

	if env == "local" {
		slog.Info("X-Ray tracer initialized in local mode (daemon may not be running)")
	} else {
		slog.Info("X-Ray tracer initialized")
	}
}
