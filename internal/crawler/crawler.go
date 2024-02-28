package crawler

import (
	"context"
	"github.com/terratensor/kremlin-parser/internal/config"
	"github.com/terratensor/kremlin-parser/internal/entities/feed"
	"github.com/terratensor/kremlin-parser/internal/lib/logger/sl"
	"github.com/terratensor/kremlin-parser/internal/parser/kremlin"
	"github.com/terratensor/kremlin-parser/internal/storage/manticore"
	"log/slog"
	"os"
	"sync"
	"time"
)

// Crawler is used as configuration for Run.
// Is validated in Run().
type Crawler struct {
	Config *config.Config
	Logger *slog.Logger
}

func (c Crawler) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	var storage feed.StorageInterface

	manticoreClient, err := manticore.New(c.Config.ManticoreIndex)
	if err != nil {
		c.Logger.Error("failed to initialize manticore client", sl.Err(err))
		os.Exit(1)
	}

	storage = manticoreClient
	entries := feed.NewFeedStorage(storage)

	for {

		for _, uri := range c.Config.StartURLs {
			prs := kremlin.NewParser(uri, c.Config, entries)
			prs.Parse(ctx, c.Logger)
		}

		select {
		case <-ctx.Done():
			break
		default:
		}

		time.Sleep(*c.Config.TimeDelay)
	}
}
