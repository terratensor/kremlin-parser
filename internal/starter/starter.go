package starter

import (
	"context"
	"github.com/terratensor/kremlin-parser/internal/config"
	"github.com/terratensor/kremlin-parser/internal/entities/feed"
	"github.com/terratensor/kremlin-parser/internal/parser/kremlin"
	"log/slog"
	"sync"
	"time"
)

// App is used as configuration for Start.
// Is validated in Start().
type App struct {
	Config *config.Config
	Logger *slog.Logger
}

func NewApp(cfg *config.Config, logger *slog.Logger) *App {
	a := &App{
		Config: cfg,
		Logger: logger,
	}
	return a
}

func (c App) Start(ctx context.Context, chout chan feed.Entry, wg *sync.WaitGroup) {
	defer wg.Done()

	kremlinCfg := c.Config.Parsers.Kremlin

	var entries []feed.Entry

	for {

		for _, uri := range kremlinCfg.StartURLs {
			prs := kremlin.NewParser(uri, 1, c.Config)
			entries = append(entries, prs.Parse(ctx, c.Logger)...)
		}

		select {
		case <-ctx.Done():
			break
		default:
		}

		for _, e := range entries {
			chout <- e
		}

		time.Sleep(*c.Config.TimeDelay)
	}
}
