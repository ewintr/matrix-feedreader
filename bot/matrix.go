package bot

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/exp/slog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"
)

type MatrixConfig struct {
	Homeserver    string
	UserID        string
	UserAccessKey string
	UserPassword  string
	RoomID        string
	DBPath        string
	Pickle        string
}

type Matrix struct {
	config       MatrixConfig
	client       *mautrix.Client
	cryptoHelper *cryptohelper.CryptoHelper
	feedReader   *Miniflux
	logger       *slog.Logger
}

func NewMatrix(cfg MatrixConfig, mflx *Miniflux, logger *slog.Logger) *Matrix {
	return &Matrix{
		config:     cfg,
		feedReader: mflx,
		logger:     logger,
	}
}

func (m *Matrix) Init() error {
	client, err := mautrix.NewClient(m.config.Homeserver, id.UserID(m.config.UserID), m.config.UserAccessKey)
	if err != nil {
		return err
	}
	var oei mautrix.OldEventIgnorer
	oei.Register(client.Syncer.(mautrix.ExtensibleSyncer))
	m.client = client
	m.cryptoHelper, err = cryptohelper.NewCryptoHelper(client, []byte(m.config.Pickle), m.config.DBPath)
	if err != nil {
		return err
	}
	m.cryptoHelper.LoginAs = &mautrix.ReqLogin{
		Type:       mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: m.config.UserID},
		Password:   m.config.UserPassword,
	}
	if err := m.cryptoHelper.Init(); err != nil {
		return err
	}
	m.client.Crypto = m.cryptoHelper

	m.AddEventHandler(m.InviteHandler())

	return nil
}

func (m *Matrix) Run() error {
	go m.PostMessages()
	go m.feedReader.Run()

	if err := m.client.Sync(); err != nil {
		return err
	}

	return nil
}

func (m *Matrix) Close() error {
	if err := m.client.Sync(); err != nil {
		return err
	}
	if err := m.cryptoHelper.Close(); err != nil {
		return err
	}

	return nil
}

func (m *Matrix) AddEventHandler(eventType event.Type, handler mautrix.EventHandler) {
	syncer := m.client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(eventType, handler)
}

func (m *Matrix) InviteHandler() (event.Type, mautrix.EventHandler) {
	return event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		if evt.GetStateKey() == m.client.UserID.String() && evt.Content.AsMember().Membership == event.MembershipInvite && evt.RoomID.String() == m.config.RoomID {
			_, err := m.client.JoinRoomByID(evt.RoomID)
			if err != nil {
				m.logger.Error("failed to join room after invite", slog.String("err", err.Error()), slog.String("room_id", evt.RoomID.String()), slog.String("inviter", evt.Sender.String()))
				return
			}

			m.logger.Info("joined room after invite", slog.String("room_id", evt.RoomID.String()), slog.String("inviter", evt.Sender.String()))
		}
	}
}

func (m *Matrix) PostMessages() {
	for entry := range m.feedReader.Feed() {
		m.logger.Info("received entry", slog.String("title", entry.Title), slog.String("url", entry.URL))
		message := fmt.Sprintf(`%s: [%s](%s)`, entry.FeedTitle, entry.Title, entry.URL)
		if entry.CommentsURL != "" {
			message += fmt.Sprintf(" - [Comments](%s)", entry.CommentsURL)
		}
		formattedMessage := format.RenderMarkdown(message, true, false)

		if _, err := m.client.SendMessageEvent(id.RoomID(m.config.RoomID), event.EventMessage, &formattedMessage); err != nil {
			m.logger.Error("failed to send message", slog.String("err", err.Error()))
			return
		}
	}
}
