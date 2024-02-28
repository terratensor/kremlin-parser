package preserver

import (
	"context"
	"github.com/terratensor/kremlin-parser/internal/config"
	"github.com/terratensor/kremlin-parser/internal/entities/feed"
	"github.com/terratensor/kremlin-parser/internal/lib/logger/sl"
	"log"
	"log/slog"
	"sync"
	"time"
)

type Preserver struct {
	Entries *feed.Entries
	Config  *config.Config
	Logger  *slog.Logger
}

func (p Preserver) Handler(ctx context.Context, ch chan feed.Entry, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case e := <-ch:

			dbe, err := p.Entries.Storage.FindByUrl(ctx, e.Url)
			if err != nil {
				p.Logger.Error("failed find entry by url", sl.Err(err))
			}
			if dbe == nil {
				id, err := p.Entries.Storage.Insert(ctx, &e)
				if err != nil {
					p.Logger.Error(
						"failed insert entry",
						slog.Int64("id", *id),
						slog.String("url", e.Url),
						sl.Err(err),
					)
				}
				p.Logger.Info(
					"entry successful inserted",
					slog.Int64("id", *id),
					slog.String("url", e.Url),
				)
			} else {
				if !matchTimes(dbe, e) {
					e.ID = dbe.ID
					err = p.Entries.Storage.Update(ctx, &e)
					if err != nil {
						p.Logger.Error(
							"failed update entry",
							slog.Int64("id", *e.ID),
							slog.String("url", e.Url),
							sl.Err(err),
						)
					} else {
						p.Logger.Info(
							"entry successful updated",
							slog.Int64("id", *e.ID),
							slog.String("url", e.Url),
						)
					}
				}
			}
		}
	}
}

func matchTimes(dbe *feed.Entry, e feed.Entry) bool {
	// Приводим время в обоих объектах к GMT+4, как на сайте Кремля
	loc, _ := time.LoadLocation("Etc/GMT-4")
	dbeTime := dbe.Updated.In(loc)
	eTime := e.Updated.In(loc)

	if dbeTime != eTime {
		log.Printf("`updated` fields do not match dbe updated %v", dbeTime)
		log.Printf("`updated` fields do not match prs updated %v", eTime)
		return false
	}
	return true
}
