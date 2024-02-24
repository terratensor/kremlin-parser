package entry

import (
	"golang.org/x/net/context"
	"time"
)

type Entry struct {
	ID        *int64     `json:"id,omitempty"`
	Language  string     `json:"language"`
	Title     string     `json:"title"`
	Url       string     `json:"url"`
	Updated   *time.Time `json:"updated"`
	Published *time.Time `json:"published"`
	Summary   string     `json:"summary"`
	Content   string     `json:"content"`
}

type StorageInterface interface {
	Insert(ctx context.Context, entry *Entry) error
	Update(ctx context.Context, entry *Entry) error
	Bulk(ctx context.Context, entries *[]Entry) error
	FindByUrl(ctx context.Context, url string) (*Entry, error)
}

type Entries struct {
	EntryStore StorageInterface
}

func NewEntries(store StorageInterface) *Entries {
	return &Entries{
		EntryStore: store,
	}
}
