package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kyberbits/forge"
)

type Config struct {
	Addr         string `env:"ADDR"`
	Port         int    `env:"PORT"`
	NotFoundFile string `env:"NOT_FOUND_FILE"`
	NotFoundCode int    `env:"NOT_FOUND_CODE"`
}

type App struct {
	config  Config
	runtime *forge.Runtime
	logger  forge.Logger
}

func (app *App) Logger() forge.Logger {
	return app.logger
}

func (app *App) Handler() http.Handler {
	fileSystem := http.Dir(".")

	static := &forge.HTTPStatic{
		FileSystem:   fileSystem,
		CacheControl: "must-revalidate, public, max-age=3600",
		NotFoundHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			file, err := fileSystem.Open(app.config.NotFoundFile)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			w.Header().Set("Cache-Control", "no-cache, no-store")
			w.WriteHeader(app.config.NotFoundCode)
			io.Copy(w, file)
		}),
	}

	return &forge.HTTPLogger{
		Logger:  app.logger,
		Handler: static,
	}
}

func (app *App) ListenAddress() string {
	return fmt.Sprintf("%s:%d", app.config.Addr, app.config.Port)
}

func (app *App) Background(ctx context.Context) {}

func main() {
	runtime := forge.NewRuntime()

	if err := runtime.ReadInDefaultEnvironmentFiles(); err != nil {
		panic(err)
	}

	config := Config{
		NotFoundFile: "404.html",
		NotFoundCode: 404,
		Port:         1234,
		Addr:         "0.0.0.0",
	}
	if err := runtime.Environment.Decode(&config); err != nil {
		panic(err)
	}

	app := &App{
		runtime: runtime,
		config:  config,
		logger: &forge.LoggerJSON{
			Encoder: json.NewEncoder(runtime.Stdout),
		},
	}

	if err := forge.Run(context.Background(), app); err != nil {
		panic(err)
	}
}
