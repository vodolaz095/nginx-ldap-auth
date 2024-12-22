//go:build !windows
// +build !windows

package zerologger

import (
	"io"
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/journald"
	"github.com/rs/zerolog/log"
)

func Configure(params Log) {
	var outputsEnabled []io.Writer

	if params.ToJournald {
		outputsEnabled = append(outputsEnabled, journald.NewJournalDWriter())
	} else {
		outputsEnabled = append(outputsEnabled, zerolog.ConsoleWriter{
			Out:        os.Stdout, // https://12factor.net/ru/logs
			TimeFormat: "15:04:05",
		})
	}
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}
	sink := zerolog.New(zerolog.MultiLevelWriter(outputsEnabled...)).
		With().Timestamp().Caller().
		Logger().Level(ExtractZerologLevel(params.Level))
	log.Logger = sink
	return
}
