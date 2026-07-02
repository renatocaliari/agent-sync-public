package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/delaneyj/toolbelt/embeddednats"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// FabricObject represents a Fabric.js canvas object
type FabricObject struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Left        float64         `json:"left"`
	Top         float64         `json:"top"`
	Width       float64         `json:"width,omitempty"`
	Height      float64         `json:"height,omitempty"`
	ScaleX      float64         `json:"scaleX,omitempty"`
	ScaleY      float64         `json:"scaleY,omitempty"`
	Angle       float64         `json:"angle,omitempty"`
	Fill        string          `json:"fill,omitempty"`
	Stroke      string          `json:"stroke,omitempty"`
	StrokeWidth float64         `json:"strokeWidth,omitempty"`
	Text        string          `json:"text,omitempty"`
	FontSize    int             `json:"fontSize,omitempty"`
	Path        json.RawMessage `json:"path,omitempty"`
	Src         string          `json:"src,omitempty"`
	Radius      float64         `json:"radius,omitempty"`
	X1          float64         `json:"x1,omitempty"`
	Y1          float64         `json:"y1,omitempty"`
	X2          float64         `json:"x2,omitempty"`
	Y2          float64         `json:"y2,omitempty"`
}

// FabricDelta represents a canvas change
type FabricDelta struct {
	Type       string        `json:"type"` // "object", "clear"
	ObjectID   string        `json:"objectId,omitempty"`
	FabricData *FabricObject `json:"fabricData,omitempty"`
	UserID     string        `json:"userId"`
	Timestamp  int64         `json:"timestamp"`
}

// FabricCanvasState represents the full canvas
type FabricCanvasState struct {
	Version string          `json:"version"`
	Objects []*FabricObject `json:"objects"`
}

type FabricService struct {
	kv jetstream.KeyValue
	nc *nats.Conn
}

func NewFabricService(ns *embeddednats.Server) (*FabricService, error) {
	nc, err := ns.Client()
	if err != nil {
		return nil, fmt.Errorf("error creating nats client: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("error creating jetstream client: %w", err)
	}

	kv, err := js.CreateOrUpdateKeyValue(context.Background(), jetstream.KeyValueConfig{
		Bucket:      "fabric_whiteboard",
		Description: "Fabric Whiteboard Objects",
		Compression: true,
		TTL:         time.Hour * 24 * 7,
		MaxBytes:    128 * 1024 * 1024, // 128MB for images
	})

	if err != nil {
		return nil, fmt.Errorf("error creating key value: %w", err)
	}

	return &FabricService{kv: kv, nc: nc}, nil
}

func (s *FabricService) GetCanvasState(ctx context.Context) (*FabricCanvasState, error) {
	keys, err := s.kv.ListKeys(ctx)
	if err != nil {
		return nil, err
	}
	defer keys.Stop()

	var objects []*FabricObject
	for key := range keys.Keys() {
		entry, err := s.kv.Get(ctx, key)
		if err != nil {
			continue
		}
		var obj FabricObject
		if err := json.Unmarshal(entry.Value(), &obj); err != nil {
			continue
		}
		objects = append(objects, &obj)
	}

	return &FabricCanvasState{
		Version: "5.3.0",
		Objects: objects,
	}, nil
}

func (s *FabricService) SaveObject(ctx context.Context, obj *FabricObject) error {
	if obj.ID == "" {
		obj.ID = fmt.Sprintf("obj_%d", time.Now().UnixNano())
	}

	key := "fabric." + obj.ID
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = s.kv.Put(ctx, key, data)
	return err
}

func (s *FabricService) DeleteObject(ctx context.Context, objectID string) error {
	key := "fabric." + objectID
	return s.kv.Delete(ctx, key)
}

func (s *FabricService) BroadcastDelta(ctx context.Context, delta *FabricDelta) {
	deltaData, _ := json.Marshal(delta)
	s.nc.Publish("fabric.whiteboard.delta", deltaData)
}

func (s *FabricService) ClearCanvas(ctx context.Context) error {
	keys, err := s.kv.ListKeys(ctx)
	if err != nil {
		return err
	}
	defer keys.Stop()

	for key := range keys.Keys() {
		s.kv.Delete(ctx, key)
	}

	delta := &FabricDelta{
		Type:      "clear",
		Timestamp: time.Now().UnixMilli(),
	}
	deltaData, _ := json.Marshal(delta)
	s.nc.Publish("fabric.whiteboard.delta", deltaData)

	return nil
}

func (s *FabricService) SubscribeDeltas(ctx context.Context) (chan *FabricDelta, error) {
	ch := make(chan *FabricDelta, 100)
	sub, err := s.nc.Subscribe("fabric.whiteboard.delta", func(m *nats.Msg) {
		var delta FabricDelta
		if err := json.Unmarshal(m.Data, &delta); err == nil {
			select {
			case ch <- &delta:
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
