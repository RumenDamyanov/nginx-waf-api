package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/RumenDamyanov/nginx-waf-api/internal/lists"
	"github.com/RumenDamyanov/nginx-waf-api/internal/reload"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	mgr      *lists.Manager
	reloader *reload.Reloader
	logger   *slog.Logger
}

// New creates a Handler.
func New(mgr *lists.Manager, reloader *reload.Reloader, logger *slog.Logger) *Handler {
	return &Handler{mgr: mgr, reloader: reloader, logger: logger}
}

// RegisterRoutes attaches handlers to the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/lists", h.listAll)
	mux.HandleFunc("GET /api/v1/lists/{name}", h.getList)
	mux.HandleFunc("POST /api/v1/lists/{name}/entries", h.addEntry)
	mux.HandleFunc("DELETE /api/v1/lists/{name}/entries/{ip...}", h.removeEntry)
	mux.HandleFunc("POST /api/v1/reload", h.triggerReload)
	mux.HandleFunc("GET /health", h.health)
}

type apiResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type addEntryRequest struct {
	IP string `json:"ip"`
}

func (h *Handler) listAll(w http.ResponseWriter, r *http.Request) {
	items, err := h.mgr.List()
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.json(w, http.StatusOK, apiResponse{Status: "ok", Data: items})
}

func (h *Handler) getList(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	detail, err := h.mgr.Get(name)
	if err != nil {
		if strings.Contains(err.Error(), "invalid list name") {
			h.error(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.Contains(err.Error(), "no such file") {
			h.error(w, http.StatusNotFound, "list not found: "+name)
			return
		}
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.json(w, http.StatusOK, apiResponse{Status: "ok", Data: detail})
}

func (h *Handler) addEntry(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	var req addEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.IP == "" {
		h.error(w, http.StatusBadRequest, "ip is required")
		return
	}

	err := h.mgr.AddEntry(name, req.IP)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.error(w, http.StatusConflict, err.Error())
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			h.error(w, http.StatusBadRequest, err.Error())
			return
		}
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.reloader.Trigger()
	h.logger.Info("entry added", "list", name, "ip", req.IP)
	h.json(w, http.StatusCreated, apiResponse{Status: "ok", Message: "entry added"})
}

func (h *Handler) removeEntry(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	ip := r.PathValue("ip")

	err := h.mgr.RemoveEntry(name, ip)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.error(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			h.error(w, http.StatusBadRequest, err.Error())
			return
		}
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.reloader.Trigger()
	h.logger.Info("entry removed", "list", name, "ip", ip)
	h.json(w, http.StatusOK, apiResponse{Status: "ok", Message: "entry removed"})
}

func (h *Handler) triggerReload(w http.ResponseWriter, r *http.Request) {
	if err := h.reloader.ReloadNow(); err != nil {
		h.error(w, http.StatusInternalServerError, "reload failed: "+err.Error())
		return
	}
	h.logger.Info("manual reload triggered")
	h.json(w, http.StatusOK, apiResponse{Status: "ok", Message: "nginx reloaded"})
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	h.json(w, http.StatusOK, apiResponse{Status: "ok"})
}

func (h *Handler) json(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) error(w http.ResponseWriter, code int, msg string) {
	h.json(w, code, apiResponse{Status: "error", Message: msg})
}
