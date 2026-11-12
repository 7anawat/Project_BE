package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"bookhub/mw"
	"bookhub/store"
)

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	who := mw.WhoFrom(r.Context())
	m, err := h.Store.MemberByID(r.Context(), who.ID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "LOGIN_REQUIRED", "กรุณาเข้าสู่ระบบก่อน")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	if err := h.populateAvatarURL(r.Context(), &m); err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

type reviewReq struct {
	Text string `json:"text"`
}

func (h *Handler) AddReview(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังสือต้องเป็นตัวเลข")
		return
	}
	var req reviewReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	text := strings.TrimSpace(req.Text)
	if utf8.RuneCountInString(text) < 10 {
		writeError(w, http.StatusBadRequest, "REVIEW_TOO_SHORT", "รีวิวสั้นเกินไป (ต้องมีอย่างน้อย 10 ตัวอักษร)")
		return
	}
	who := mw.WhoFrom(r.Context())
	rv, err := h.Store.AddReview(r.Context(), who.ID, bookID, text)
	switch {
	case errors.Is(err, store.ErrDuplicate):
		writeError(w, http.StatusConflict, "ALREADY_REVIEWED", "คุณรีวิวหนังสือเล่มนี้ไปแล้ว")
		return
	case errors.Is(err, store.ErrNoBook):
		writeError(w, http.StatusNotFound, "BOOK_NOT_FOUND", "ไม่พบหนังสือเล่มนี้")
		return
	case err != nil:
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": rv.ID, "text": rv.Text, "createdAt": rv.CreatedAt})
}

func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	reviewID, ok := pathID(r, "reviewId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขรีวิวต้องเป็นตัวเลข")
		return
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.DeleteReview(r.Context(), reviewID, who.ID, who.Role == "admin")
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "REVIEW_NOT_FOUND", "ไม่พบรีวิวนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetWishlist(w http.ResponseWriter, r *http.Request) {
	who := mw.WhoFrom(r.Context())
	items, err := h.Store.GetWishlist(r.Context(), who.ID)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) AddToWishlist(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(r, "bookId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังสือต้องเป็นตัวเลข")
		return
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.AddToWishlist(r.Context(), who.ID, bookID)
	switch {
	case errors.Is(err, store.ErrAlreadyInWishlist):
		writeError(w, http.StatusConflict, "ALREADY_IN_WISHLIST", "หนังสือเล่มนี้อยู่ในรายการอยากอ่านแล้ว")
		return
	case errors.Is(err, store.ErrNoBook):
		writeError(w, http.StatusNotFound, "BOOK_NOT_FOUND", "ไม่พบหนังสือเล่มนี้")
		return
	case err != nil:
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveFromWishlist(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(r, "bookId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังสือต้องเป็นตัวเลข")
		return
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.RemoveFromWishlist(r.Context(), who.ID, bookID)
	if errors.Is(err, store.ErrNotFound) {
		w.WriteHeader(http.StatusNoContent) // ไม่มีก็ถือว่าลบแล้ว
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
