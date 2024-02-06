package main

import (
	"github.com/terratensor/kremlin-parser/internal/config"
	"github.com/terratensor/kremlin-parser/internal/lib/logger/handlers/slogpretty"
	"github.com/terratensor/kremlin-parser/internal/lib/logger/sl"
	"github.com/terratensor/kremlin-parser/internal/parser"
	"github.com/terratensor/kremlin-parser/internal/storage/sqlite"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env)) // к каждому сообщению будет добавляться поле с информацией о текущем окружении

	log.Debug("logger debug mode enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
		os.Exit(1)
	}

	//var pageCount, outputPath string
	//
	//flag.StringVarP(&cfg.Parser.OutputPath, "output", "o", "./data", "путь сохранения файлов")
	//flag.StringVarP(&cfg.Parser.PageCount, "page-count", "p", "1", "спарсить указанное количество страниц")
	//flag.Parse()

	prs := parser.New(cfg, storage)
	prs.Parse(log)
	log.Info("all pages were successfully parsed")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
