package main

import (
	"context"
	flag "github.com/spf13/pflag"
	"github.com/terratensor/kremlin-parser/internal/config"
	"github.com/terratensor/kremlin-parser/internal/crawler"
	"github.com/terratensor/kremlin-parser/internal/entities/feed"
	"github.com/terratensor/kremlin-parser/internal/lib/logger/handlers/slogpretty"
	"github.com/terratensor/kremlin-parser/internal/lib/logger/sl"
	"github.com/terratensor/kremlin-parser/internal/parser/kremlin"
	"github.com/terratensor/kremlin-parser/internal/storage/manticore"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	prepareTimeZone()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	cfg := config.MustLoad()

	logger := setupLogger(cfg.Env)
	logger = logger.With(slog.String("env", cfg.Env))
	logger.Debug("logger debug mode enabled")

	var demon bool
	var pageCount int

	flag.BoolVarP(&demon, "service", "s", false, "запуск парсера в режиме службы")
	flag.IntVarP(&pageCount, "page-count", "p", 0, "спарсить указанное количество страниц")
	flag.Parse()

	var storage feed.StorageInterface

	manticoreClient, err := manticore.New(cfg.ManticoreIndex)
	if err != nil {
		logger.Error("failed to initialize manticore client", sl.Err(err))
		os.Exit(1)
	}

	storage = manticoreClient
	entries := feed.NewFeedStorage(storage)

	if demon {
		//ch := make(chan feed.Entry, 100)
		wg := &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			crawler.Crawler{
				Config: cfg,
				Logger: logger,
			}.Run(ctx, wg)
			// Обрабатываем ошибку и выходим с кодом 1, для того чтобы инициировать перезапуск докер контейнера.
			// Возможно тут имеет смысл сделать сервис проверки health, но пока так
			//if err != nil {
			//	logger.Error("%v\r\n failure, restart required", sl.Err(err))
			//	//sentry.CaptureMessage(fmt.Sprint(err))
			//	os.Exit(1)
			//}
		}()
		wg.Wait()
		cancel()
	}

	if pageCount > 0 {
		cfg.PageCount = pageCount
	}

	for _, uri := range cfg.StartURLs {
		prs := kremlin.NewParser(uri, cfg, entries)
		prs.Parse(ctx, logger)
	}
	logger.Info("all pages were successfully parsed")
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
