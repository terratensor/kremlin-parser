package entry

import (
	"golang.org/x/net/context"
	"time"
)

//type Entries []Entry

type Entry struct {
	Title     string     `json:"title"`
	Url       string     `json:"url"`
	Updated   *time.Time `json:"update"`
	Published *time.Time `json:"published"`
	Content   string     `json:"content"`
}

type Storage interface {
	Create(ctx context.Context, entry *Entry) error
	BatchSave(ctx context.Context, entries Entries) error
}

type Entries struct {
	EntryStore Storage
}

func NewEntries(store Storage) *Entries {
	return &Entries{
		EntryStore: store,
	}
}

func (e *Entry) Create(entry *Entry) error {
	return nil
}
