package log

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func init() {
	Setup()
}

var SubLog zerolog.Logger

func Setup() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	SubLog = log.Output(
		zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: false,
		},
	)

	SubLog.Info().Msg("Logger started")
}

func Info(log string, v ...interface{}) {
	SubLog.Info().Msg(fmt.Sprintf(log, v...))
}

func Debug(log string, v ...interface{}) {
	SubLog.Debug().Msg(fmt.Sprintf(log, v...))
}

func Warn(log string, v ...interface{}) {
	SubLog.Warn().Msg(fmt.Sprintf(log, v...))
}

func Error(log string, v ...interface{}) {
	SubLog.Error().Msg(fmt.Sprintf(log, v...))
}
