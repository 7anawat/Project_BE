package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"bookhub/auth"
	"bookhub/store"
)

type Handler struct {
	Store   *store.Store
	Tickets *auth.Issuer
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

// internal จดรายละเอียดไว้ใน log ฝั่งเรา แต่บอกผู้ใช้แค่ว่าระบบขัดข้อง
func internal(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	writeError(w, http.StatusInternalServerError, "INTERNAL", "ระบบขัดข้อง ลองใหม่อีกครั้ง")
}

func pathID(r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	return id, err == nil && id > 0
}

func (h *Handler) SearchBooks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	items, err := h.Store.SearchBooks(r.Context(), q, category)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังสือต้องเป็นตัวเลข")
		return
	}
	b, err := h.Store.GetBook(r.Context(), bookID)
	if err != nil {
		if err.Error() == "not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "ไม่พบหนังสือเล่มนี้")
			return
		}
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListCategories(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังสือต้องเป็นตัวเลข")
		return
	}
	items, err := h.Store.ListReviews(r.Context(), bookID)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

