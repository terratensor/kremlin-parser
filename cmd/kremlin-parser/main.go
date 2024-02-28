package main

import (
	"context"
	"github.com/terratensor/kremlin-parser/internal/config"
	"github.com/terratensor/kremlin-parser/internal/entities/feed"
	"github.com/terratensor/kremlin-parser/internal/lib/logger/handlers/slogpretty"
	"github.com/terratensor/kremlin-parser/internal/lib/logger/sl"
	"github.com/terratensor/kremlin-parser/internal/parser"
	"github.com/terratensor/kremlin-parser/internal/storage/manticore"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	prepareTimeZone()
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env)) // к каждому сообщению будет добавляться поле с информацией о текущем окружении

	log.Debug("logger debug mode enabled")

	var storage feed.StorageInterface

	manticoreClient, err := manticore.New(cfg.ManticoreIndex)
	if err != nil {
		log.Error("failed to initialize manticore client", sl.Err(err))
		os.Exit(1)
	}

	storage = manticoreClient
	entries := feed.NewFeedStorage(storage)

	//var pageCount, outputPath string
	//
	//flag.StringVarP(&cfg.Parser.OutputPath, "output", "o", "./data", "путь сохранения файлов")
	//flag.StringVarP(&cfg.Parser.PageCount, "page-count", "p", "1", "спарсить указанное количество страниц")
	//flag.Parse()

	//var wg sync.WaitGroup
	//for _, uri := range cfg.StartURLs {
	//	prs := parser.New(uri, cfg, manticoreClient)
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		prs.Parse(log)
	//	}()
	//}
	//wg.Wait()

	for _, uri := range cfg.StartURLs {
		prs := parser.New(uri, cfg, entries)
		prs.Parse(ctx, log)
	}

	log.Debug("all pages were successfully parsed")
}

// setupLogger инициализирует и возвращает logger в зависимости от окружения.
//
// Принимает строковый параметр, представляющий среду, и возвращает указатель на slog.Logger.
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

func prepareTimeZone() {
	if tz := os.Getenv("TZ"); tz != "" {
		var err error
		time.Local, err = time.LoadLocation(tz)
		if err != nil {
			log.Printf("error loading location '%s': %v\n", tz, err)
		}
	}

	// output current time zone
	tnow := time.Now()
	tz, _ := tnow.Zone()
	log.Printf("Local time zone %s. Service started at %s", tz,
		tnow.Format("2006-01-02T15:04:05.000 MST"))
}
