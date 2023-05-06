package main

import (
	"golang.org/x/exp/slog"
)

func main() {
	config, err := ReadConfig()
	if err != nil {
		slog.Error("there was problem when reading config, stop execution", err)
		panic(err)
	}

	err = writeMessage(config)
	if err != nil {
		slog.Error("there was problem when writing message, stop execution", err)
		panic(err)
	}

	slog.Info("finished")
}
