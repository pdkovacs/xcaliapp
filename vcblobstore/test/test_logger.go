package test

import (
	"os"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

func createTestLogger() zerolog.Logger {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	return zerolog.New(output).
		Level(zerolog.Level(zerolog.DebugLevel)).
		With().
		Str("app_xid", xid.New().String()).
		Caller().
		Logger()
}
