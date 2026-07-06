package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/k22mitta/pocket-api/internal/ai"
	"github.com/k22mitta/pocket-api/internal/api/middleware"
	"github.com/k22mitta/pocket-api/internal/db/queries"
	"github.com/k22mitta/pocket-api/internal/models"
)

func Chat(aiClient *ai.Client, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		var body struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if body.Message == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message is required"})
			return
		}

		systemPrompt, err := ai.BuildFinancialContext(db, userID)
		if err != nil {
			log.Printf("chat error: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		history, err := queries.GetChatHistory(db, userID, 20)
		if err != nil {
			log.Printf("chat error: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		aiHistory := make([]ai.ChatMessage, len(history))
		for i, m := range history {
			aiHistory[i] = ai.ChatMessage{Role: m.Role, Content: m.Content}
		}

		if err := queries.SaveChatMessage(db, userID, "user", body.Message); err != nil {
			log.Printf("chat error: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		response, err := aiClient.Chat(r.Context(), systemPrompt, body.Message, aiHistory)
		if err != nil {
			log.Printf("chat error: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		if err := queries.SaveChatMessage(db, userID, "assistant", response); err != nil {
			log.Printf("chat error: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"response": response,
			"message":  body.Message,
		})
	}
}

func GetChatHistory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())

		history, err := queries.GetChatHistory(db, userID, 50)
		if err != nil {
			log.Printf("chat error: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}

		if history == nil {
			history = []models.ChatMessage{}
		}

		writeJSON(w, http.StatusOK, history)
	}
}
