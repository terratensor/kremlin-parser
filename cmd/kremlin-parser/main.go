package main

import (
	"context"
	flag "github.com/spf13/pflag"
	"github.com/terratensor/kremlin-parser/internal/config"
	"github.com/terratensor/kremlin-parser/internal/entities/feed"
	"github.com/terratensor/kremlin-parser/internal/lib/logger/handlers/slogpretty"
	"github.com/terratensor/kremlin-parser/internal/lib/logger/sl"
	"github.com/terratensor/kremlin-parser/internal/parser/kremlin"
	"github.com/terratensor/kremlin-parser/internal/preserver"
	"github.com/terratensor/kremlin-parser/internal/starter"
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

	var service bool
	var pageCount int

	flag.BoolVarP(&service, "service", "s", false, "запуск парсера в режиме службы")
	flag.IntVarP(&pageCount, "page-count", "p", 1, "спарсить указанное количество страниц")
	flag.Parse()

	var storage feed.StorageInterface

	manticoreClient, err := manticore.New(cfg.ManticoreIndex)
	if err != nil {
		logger.Error("failed to initialize manticore client", sl.Err(err))
		os.Exit(1)
	}

	storage = manticoreClient

	ch := make(chan feed.Entry)
	wg := &sync.WaitGroup{}

	if service {

		app := starter.NewApp(cfg, logger)

		wg.Add(2)

		go app.Start(ctx, ch, wg)

		go preserver.Preserver{
			Entries: feed.NewFeedStorage(storage),
			Logger:  logger,
			Config:  cfg,
		}.Handler(ctx, ch, wg)

		wg.Wait()
		cancel()
	}

	// Если запущено ни как служба и задано число страниц
	// TODO это надо переделать, т.к. не будет работать если парсеров будет больше чем 1
	// Надо реализовать завершение, похоже канал не закрывается? Разобраться с каналом
	logger.Info("app mode page parser")

	wg.Add(1)

	var entries []feed.Entry
	for _, uri := range cfg.Parsers.Kremlin.StartURLs {
		prs := kremlin.NewParser(uri, pageCount, cfg)
		entries = append(entries, prs.Parse(ctx, logger)...)
	}

	go preserver.Preserver{
		Entries: feed.NewFeedStorage(storage),
		Logger:  logger,
		Config:  cfg,
	}.Handler(ctx, ch, wg)

	for _, e := range entries {
		ch <- e
	}

	wg.Wait()
	logger.Info("all pages were successfully parsed")
	cancel()
	<-ctx.Done()

}

// setupLogger инициализирует и возвращает logger в зависимости от окружения.
//
// Принимает строковый параметр, представляющий среду, и возвращает указатель на slog.Logger.
func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = setupPrettySlog()
	case envDev:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return logger
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
