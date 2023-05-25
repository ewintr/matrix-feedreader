package main

import (
	"os"
	"os/signal"

	"ewintr.nl/matrix-feedreader/bot"
	"golang.org/x/exp/slog"
)

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	mflxConfig := bot.MinifluxInfo{
		Endpoint: getParam("MINIFLUX_ENDPOINT", "http://localhost:8080"),
		ApiKey:   getParam("MINIFLUX_API_KEY", "secret"),
	}
	mflx := bot.NewMiniflux(mflxConfig, logger)

	mtrxConf := bot.MatrixConfig{
		Homeserver:    getParam("MATRIX_HOMESERVER", "http://localhost/"),
		UserID:        getParam("MATRIX_USER_ID", "@user:localhost"),
		UserAccessKey: getParam("MATRIX_USER_ACCESS_KEY", "secret"),
		UserPassword:  getParam("MATRIX_USER_PASSWORD", "secret"),
		RoomID:        getParam("MATRIX_ROOM_ID", "!room:localhost"),
		DBPath:        getParam("MATRIX_DB_PATH", "matrix.db"),
		Pickle:        getParam("MATRIX_PICKLE", "matrix.pickle"),
	}
	mtrx := bot.NewMatrix(mtrxConf, mflx, logger)
	if err := mtrx.Init(); err != nil {
		logger.Error("error running matrix bot: %v", slog.String("error", err.Error()))
		os.Exit(1)
	}

	go mtrx.Run()
	logger.Info("matrix bot started")

	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt)
	<-done

	logger.Info("matrix bot stopped")
}

func getParam(param, def string) string {
	if val, ok := os.LookupEnv(param); ok {
		return val
	}
	return def
}
