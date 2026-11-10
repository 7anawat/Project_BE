package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"bookhub/mw"
	"bookhub/store"
)

func (h *Handler) AdminMembers(w http.ResponseWriter, r *http.Request) {
	ms, err := h.Store.ListMembers(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": ms})
}

func (h *Handler) AdminSetRole(w http.ResponseWriter, r *http.Request) {
	memberID, ok := pathID(r, "memberId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขสมาชิกต้องเป็นตัวเลข")
		return
	}
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	role := strings.TrimSpace(req.Role)
	if role != "member" && role != "admin" {
		writeError(w, http.StatusBadRequest, "BAD_ROLE", "role ต้องเป็น member หรือ admin เท่านั้น")
		return
	}
	who := mw.WhoFrom(r.Context())
	if who.ID == memberID {
		writeError(w, http.StatusForbidden, "SELF_ROLE", "ไม่สามารถเปลี่ยนบทบาทของตัวเองได้")
		return
	}
	m, err := h.Store.SetMemberRole(r.Context(), memberID, role)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "MEMBER_NOT_FOUND", "ไม่พบสมาชิกนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *Handler) AdminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.Store.GetAdminStats(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

type bookPayload struct {
	Title           string   `json:"title"`
	Author          string   `json:"author"`
	Category        *string  `json:"category"`
	ISBN            *string  `json:"isbn"`
	Publisher       *string  `json:"publisher"`
	PublicationYear *int     `json:"publicationYear"`
	Description     *string  `json:"description"`
	PriceTHB        *float64 `json:"priceTHB"`
	StockQuantity   int      `json:"stockQuantity"`
	CoverURL        *string  `json:"coverUrl"`
	Language        string   `json:"language"`
}

func (h *Handler) AdminCreateBook(w http.ResponseWriter, r *http.Request) {
	var req bookPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "BAD_TITLE", "ต้องมีชื่อหนังสือ")
		return
	}
	if strings.TrimSpace(req.Author) == "" {
		writeError(w, http.StatusBadRequest, "BAD_AUTHOR", "ต้องมีชื่อผู้แต่ง")
		return
	}
	lang := req.Language
	if lang == "" {
		lang = "th"
	}
	b, err := h.Store.CreateBook(r.Context(), store.BookInput{
		Title: req.Title, Author: req.Author, Category: req.Category,
		ISBN: req.ISBN, Publisher: req.Publisher, PublicationYear: req.PublicationYear,
		Description: req.Description, PriceTHB: req.PriceTHB,
		StockQuantity: req.StockQuantity, CoverURL: req.CoverURL, Language: lang,
	})
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) AdminGetBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังสือต้องเป็นตัวเลข")
		return
	}
	b, err := h.Store.GetBook(r.Context(), bookID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "ไม่พบหนังสือเล่มนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) AdminUpdateBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังสือต้องเป็นตัวเลข")
		return
	}
	var req bookPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "BAD_TITLE", "ต้องมีชื่อหนังสือ")
		return
	}
	lang := req.Language
	if lang == "" {
		lang = "th"
	}
	b, err := h.Store.UpdateBook(r.Context(), bookID, store.BookInput{
		Title: req.Title, Author: req.Author, Category: req.Category,
		ISBN: req.ISBN, Publisher: req.Publisher, PublicationYear: req.PublicationYear,
		Description: req.Description, PriceTHB: req.PriceTHB,
		StockQuantity: req.StockQuantity, CoverURL: req.CoverURL, Language: lang,
	})
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "ไม่พบหนังสือเล่มนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) AdminDeleteBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังสือต้องเป็นตัวเลข")
		return
	}
	err := h.Store.DeleteBook(r.Context(), bookID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "ไม่พบหนังสือเล่มนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

