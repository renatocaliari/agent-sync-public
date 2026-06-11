package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/delaneyj/toolbelt/embeddednats"
	"github.com/nats-io/nats.go"
)

type Cursor struct {
	UserID string  `json:"userId"`
	Name   string  `json:"name"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Color  string  `json:"color"`
}

var funnyNames = []string{
	"Pincel Malandro", "Desenhista Doideira", "Artista Descarado",
	"Risco Radiante", "Garatuja Genial", "Pintor Pitoco",
	"Caneta Criativa", "Borracha Bizarra", "Traço Trapalhão",
	"Forma Fantástica", "Texto Trapaceiro", "Cursor Cômico",
	"Artista Amador", "Desenho Doido", "Pintor Preguiçoso",
	"Rabisco Raivoso", "Esboço Esperto", "Traço Tremido",
}

var colors = []string{"#ef4444", "#3b82f6", "#22c55e", "#f59e0b"}

type CursorService struct {
	nc    *nats.Conn
	mu    sync.RWMutex
	users map[string]*Cursor
}

func NewCursorService(ns *embeddednats.Server) (*CursorService, error) {
	nc, err := ns.Client()
	if err != nil {
		return nil, fmt.Errorf("error creating nats client: %w", err)
	}

	return &CursorService{
		nc:    nc,
		users: make(map[string]*Cursor),
	}, nil
}

func (s *CursorService) GetOrCreateCursor(userID string) *Cursor {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.users[userID]; ok {
		return c
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	c := &Cursor{
		UserID: userID,
		Name:   funnyNames[r.Intn(len(funnyNames))],
		Color:  colors[r.Intn(len(colors))],
	}
	s.users[userID] = c
	return c
}

func (s *CursorService) UpdateCursor(userID string, x, y float64) error {
	s.mu.Lock()
	c, ok := s.users[userID]
	if !ok {
		// Create cursor if doesn't exist
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		c = &Cursor{
			UserID: userID,
			Name:   funnyNames[r.Intn(len(funnyNames))],
			Color:  colors[r.Intn(len(colors))],
		}
		s.users[userID] = c
	}
	c.X = x
	c.Y = y
	cursorCopy := *c
	s.mu.Unlock()

	data, err := json.Marshal(&cursorCopy)
	if err != nil {
		return err
	}
	return s.nc.Publish("whiteboard.cursors", data)
}

func (s *CursorService) Subscribe(ctx context.Context) (chan *Cursor, error) {
	ch := make(chan *Cursor, 100)
	sub, err := s.nc.Subscribe("whiteboard.cursors", func(m *nats.Msg) {
		var c Cursor
		if err := json.Unmarshal(m.Data, &c); err == nil {
			select {
			case ch <- &c:
			default:
			}
		}
	})
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		sub.Unsubscribe()
		close(ch)
	}()

	return ch, nil
}

func (s *CursorService) GetAllCursors() []*Cursor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cursors := make([]*Cursor, 0, len(s.users))
	for _, c := range s.users {
		cursors = append(cursors, c)
	}
	return cursors
}
