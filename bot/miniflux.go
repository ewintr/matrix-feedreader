package bot

import (
	"miniflux.app/client"
)

type Entry struct {
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
}

func NewMiniflux(mflInfo MinifluxInfo) *Miniflux {
	return &Miniflux{
		client: client.New(mflInfo.Endpoint, mflInfo.ApiKey),
	}
}

func (m *Miniflux) Unread() ([]Entry, error) {
	result, err := m.client.Entries(&client.Filter{Status: "unread"})
	if err != nil {
		return nil, err
	}

	entries := []Entry{}
	for _, entry := range result.Entries {
		entries = append(entries, Entry{
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
