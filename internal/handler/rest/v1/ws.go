package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"go.uber.org/zap"
)

type DocumentSyncer interface {
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

		h.logger.Info("ws message", zap.Any("message", msg))
	}
}
