package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	dmp "github.com/sergi/go-diff/diffmatchpatch"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type DocumentSyncer interface {
	Update(ctx context.Context, id string, diffs []dmp.Diff)
}

type WSHandler struct {
	upgrader       websocket.Upgrader
	documentSyncer DocumentSyncer
	logger         *logger.ZapLogger
}

func NewWSHandler(documentSyncer DocumentSyncer, logger *logger.ZapLogger) *WSHandler {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return &WSHandler{
		upgrader:       upgrader,
		documentSyncer: documentSyncer,
		logger:         logger,
	}
}

type DiffMessage struct {
	DocumentID string `json:"documentId"`
	Version    int    `json:"version"`
	Diff       []struct {
		Operation string `json:"op"`
		Line      int    `json:"line"`
		Text      string `json:"text"`
	} `json:"diff"`
}

func (m DiffMessage) ToDiff() ([]dmp.Diff, error) {
	var diffs []dmp.Diff
	for _, d := range m.Diff {
		diffs = append(diffs, dmp.Diff{
			Text: d.Text,
		})
	}

	return diffs, nil
}

func (h *WSHandler) Handle(w http.ResponseWriter, r *http.Request) {
	conn, connErr := h.upgrader.Upgrade(w, r, nil)
	if connErr != nil {
		http.Error(w, "could not upgrade", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			h.logger.Error("conn.ReadMessage", zap.Error(err))
			break
		}

		var msg DiffMessage
		err = json.Unmarshal(msgBytes, &msg)
		if err != nil {
			h.logger.Error("json.Unmarshal", zap.Error(err))
			continue
		}

		diffs, err := msg.ToDiff()
		if err != nil {
			h.logger.Error("msg.ToDiff", zap.Error(err))
			continue
		}

		h.documentSyncer.Update(r.Context(), msg.DocumentID, diffs)
		h.logger.Info("ws message", zap.Any("message", msg))
	}
}
