package entry

import (
	"golang.org/x/net/context"
	"time"
)

//type Storage []Entry

type Entry struct {
	Language  string     `json:"language"`
	Title     string     `json:"title"`
	Url       string     `json:"url"`
	Updated   *time.Time `json:"updated"`
	Published *time.Time `json:"published"`
	Summary   string     `json:"summary"`
	Content   string     `json:"content"`
}

type StorageInterface interface {
	Insert(ctx context.Context, entry *Entry)
	Bulk(ctx context.Context, entries *[]Entry)
}

type Entries struct {
	EntryStore StorageInterface
}

func NewEntries(store StorageInterface) *Entries {
	return &Entries{
		EntryStore: store,
	}
}

func (e *Entry) Insert(entry *Entry) error {
	return nil
}
