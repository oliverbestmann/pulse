package main

import (
	_ "image/jpeg"
	"log/slog"
	"os"

	"github.com/oliverbestmann/go3d/orion"
)

func main() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	opts := orion.RunGameOptions{
		Game: &TestGame{},
	}

	if err := orion.RunGame(opts); err != nil {
		panic(err)
	}
}
