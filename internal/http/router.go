package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	kafka "github.com/segmentio/kafka-go"

	"wb-order-service/internal/repo"
	"wb-order-service/internal/service"
)

type Router struct {
	r     *chi.Mux
	l     *slog.Logger
	s     *service.Service
	pg    *pgxpool.Pool
	kb    []string
	topic string
}

func New(s *service.Service, logger *slog.Logger, pg *pgxpool.Pool, kafkaBrokers []string, topic string) *Router {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Recoverer,
		middleware.Timeout(15*time.Second),
	)

	h := &Router{
		r:     r,
		l:     logger,
		s:     s,
		pg:    pg,
		kb:    kafkaBrokers,
		topic: topic,
	}

	r.Get("/healthz", h.healthz)
	r.Get("/readyz", h.readyz)
	r.Get("/order/{id}", h.getOrder)
	r.Handle("/", http.FileServer(http.Dir("./web")))

	return h
}

func (h *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.r.ServeHTTP(w, r) }

func (h *Router) healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *Router) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// DB
	if err := h.pg.Ping(ctx); err != nil {
		http.Error(w, "db not ready", http.StatusServiceUnavailable)
		return
	}

	// Kafka (опциональная проверка лидера; если нет брокеров — пропускаем)
	kafkaStatus := "skipped"
	if len(h.kb) > 0 {
		kafkaStatus = "checked"
		if c, err := kafka.DialLeader(ctx, "tcp", h.kb[0], h.topic, 0); err == nil {
			_ = c.Close()
		} else {
			// Не валим readiness, но логируем
			h.l.Warn("readyz: kafka check failed", "err", err)
			kafkaStatus = "unavailable"
		}
	}

	resp := map[string]any{"db": "ok", "kafka": kafkaStatus}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(resp); err != nil {
		h.l.Error("readyz: encode response", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = buf.WriteTo(w)
}

func (h *Router) getOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" || len(id) < 6 {
		http.Error(w, "bad order id", http.StatusBadRequest)
		return
	}

	o, err := h.s.GetOrder(r.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.l.Error("getOrder: service error", "id", id, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(o); err != nil {
		h.l.Error("getOrder: encode response", "id", id, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = buf.WriteTo(w)
}

