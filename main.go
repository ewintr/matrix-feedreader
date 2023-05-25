package main

import (
	"os"

	"ewintr.nl/matrix-feedreader/bot"
	"golang.org/x/exp/slog"
)

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mflx := bot.NewMiniflux(bot.MinifluxInfo{
		Endpoint: getParam("MINIFLUX_ENDPOINT", "http://localhost:8080"),
		ApiKey:   getParam("MINIFLUX_APIKEY", "secret"),
	})

	unread, err := mflx.Unread()
	if err != nil {
		logger.Error("could not get unread", slog.String("err", err.Error()))
		return
	}

	for _, entry := range unread {
		logger.Info("entry", slog.String("title", entry.Title))
	}
}

func getParam(param, def string) string {
	if val, ok := os.LookupEnv(param); ok {
		return val
	}
	return def
}
