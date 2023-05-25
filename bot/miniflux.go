package bot

import (
	"time"

	"golang.org/x/exp/slog"
	"miniflux.app/client"
)

type Entry struct {
	ID          int64
	Title       string
	Description string
	URL         string
	Comments    string
	FeedTitle   string
}

type MinifluxInfo struct {
	Endpoint string
	ApiKey   string
}

type Miniflux struct {
	client *client.Client
	logger *slog.Logger
	feed   chan Entry
}

func NewMiniflux(mflInfo MinifluxInfo, logger *slog.Logger) *Miniflux {
	return &Miniflux{
		client: client.New(mflInfo.Endpoint, mflInfo.ApiKey),
		logger: logger,
		feed:   make(chan Entry),
	}
}

func (m *Miniflux) Run() {
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-ticker.C:
			entries, err := m.Unread()
			if err != nil {
				m.logger.Error("error getting unread entries", slog.String("error", err.Error()))
				continue
			}

			for _, entry := range entries {
				m.feed <- entry
				m.MarkRead(entry.ID)
			}
		}
	}
}

func (m *Miniflux) Feed() <-chan Entry {
	return m.feed
}

func (m *Miniflux) Unread() ([]Entry, error) {
	result, err := m.client.Entries(&client.Filter{Status: "unread"})
	if err != nil {
		return nil, err
	}

	entries := []Entry{}
	for _, entry := range result.Entries {
		entries = append(entries, Entry{
			ID:          entry.ID,
			Title:       entry.Title,
			Description: entry.Content,
			URL:         entry.URL,
			Comments:    entry.CommentsURL,
			FeedTitle:   entry.Feed.Title,
		})
	}

	return entries, nil
}

func (m *Miniflux) MarkRead(entryID int64) error {
	if err := m.client.UpdateEntries([]int64{entryID}, "read"); err != nil {
		return err
	}

	return nil
}
