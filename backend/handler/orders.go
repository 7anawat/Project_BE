package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"bookhub/mw"
	"bookhub/store"
)

const maxSlipSize = 5 << 20 // 5 MB

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	memberID := mw.WhoFrom(r.Context()).ID
	if memberID <= 0 {
		writeError(w, http.StatusUnauthorized, "LOGIN_REQUIRED", "กรุณาเข้าสู่ระบบ")
		return
	}
	if h.Storage == nil {
		internal(w, errors.New("storage client is not configured"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxSlipSize+(1<<20))
	if err := r.ParseMultipartForm(maxSlipSize); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_FORM", "ข้อมูลหรือไฟล์สลิปไม่ถูกต้อง หรือมีขนาดใหญ่เกินไป")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	bookIDsText := r.FormValue("bookIds")
	var bookIDs []int64
	if err := json.Unmarshal([]byte(bookIDsText), &bookIDs); err != nil || len(bookIDs) == 0 {
		writeError(w, http.StatusBadRequest, "INVALID_BOOKS", "กรุณาเลือกรายการหนังสือให้ถูกต้อง")
		return
	}

	file, header, err := r.FormFile("slip")
	if err != nil {
		writeError(w, http.StatusBadRequest, "SLIP_REQUIRED", "กรุณาแนบรูปสลิป")
		return
	}
	defer file.Close()

	if header.Size > maxSlipSize {
		writeError(w, http.StatusBadRequest, "SLIP_TOO_LARGE", "ไฟล์สลิปต้องมีขนาดไม่เกิน 5 MB")
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, maxSlipSize+1))
	if err != nil {
		internal(w, err)
		return
	}
	if len(data) == 0 || len(data) > maxSlipSize {
		writeError(w, http.StatusBadRequest, "INVALID_SLIP", "ไฟล์สลิปว่างหรือมีขนาดใหญ่เกินไป")
		return
	}

	contentType := http.DetectContentType(data)
	extension := ""
	switch contentType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/webp":
		extension = ".webp"
	default:
		writeError(w, http.StatusBadRequest, "INVALID_SLIP_TYPE", "รองรับเฉพาะไฟล์ JPG, PNG และ WEBP")
		return
	}

	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		internal(w, err)
		return
	}

	objectPath := fmt.Sprintf("%d/%s%s", memberID, hex.EncodeToString(randomBytes), extension)

	if err := h.Storage.Upload(r.Context(), objectPath, data, contentType); err != nil {
		internal(w, err)
		return
	}

	order, err := h.Store.CreateOrder(
		r.Context(),
		memberID,
		bookIDs,
		objectPath,
		strings.TrimSpace(r.FormValue("transferBank")),
		strings.TrimSpace(r.FormValue("slipNote")),
	)
	if err != nil {
		_ = h.Storage.Remove(r.Context(), objectPath)

		switch {
		case errors.Is(err, store.ErrInvalidOrder):
			writeError(w, http.StatusBadRequest, "INVALID_ORDER", "ข้อมูลคำสั่งซื้อไม่ถูกต้อง")
		case errors.Is(err, store.ErrNotFound):
			writeError(w, http.StatusNotFound, "BOOK_NOT_FOUND", "ไม่พบหนังสือบางรายการ")
		default:
			internal(w, err)
		}
		return
	}

	signedURL, err := h.Storage.SignedURL(r.Context(), objectPath)
	if err != nil {
		internal(w, err)
		return
	}
	order.SlipURL = signedURL

	writeJSON(w, http.StatusCreated, map[string]any{"order": order})
}

func (h *Handler) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	memberID := mw.WhoFrom(r.Context()).ID

	orders, err := h.Store.ListMemberOrders(r.Context(), memberID)
	if err != nil {
		internal(w, err)
		return
	}

	if err := h.attachSlipURLs(r, orders); err != nil {
		internal(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": orders})
}

func (h *Handler) AdminOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.Store.ListAllOrders(r.Context())
	if err != nil {
		internal(w, err)
		return
	}

	if err := h.attachSlipURLs(r, orders); err != nil {
		internal(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": orders})
}

func (h *Handler) attachSlipURLs(r *http.Request, orders []store.Order) error {
	if h.Storage == nil {
		return errors.New("storage client is not configured")
	}

	for i := range orders {
		if orders[i].SlipPath == "" {
			continue
		}
		signedURL, err := h.Storage.SignedURL(r.Context(), orders[i].SlipPath)
		if err != nil {
			return err
		}
		orders[i].SlipURL = signedURL
	}
	return nil
}

func (h *Handler) AdminApproveOrder(w http.ResponseWriter, r *http.Request) {
	orderID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขคำสั่งซื้อไม่ถูกต้อง")
		return
	}

	adminID := mw.WhoFrom(r.Context()).ID
	memberID, err := h.Store.ApproveOrder(r.Context(), orderID, adminID)
	if err != nil {
		h.writeOrderError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "approved",
		"memberId": memberID,
		"orderId":  orderID,
	})
}

func (h *Handler) AdminRejectOrder(w http.ResponseWriter, r *http.Request) {
	orderID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขคำสั่งซื้อไม่ถูกต้อง")
		return
	}

	var input struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "ข้อมูลการปฏิเสธไม่ถูกต้อง")
		return
	}

	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		writeError(w, http.StatusBadRequest, "REASON_REQUIRED", "กรุณาระบุเหตุผลที่ปฏิเสธ")
		return
	}

	adminID := mw.WhoFrom(r.Context()).ID
	if err := h.Store.RejectOrder(r.Context(), orderID, adminID, reason); err != nil {
		h.writeOrderError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "rejected",
		"orderId": orderID,
	})
}

func (h *Handler) writeOrderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "ORDER_NOT_FOUND", "ไม่พบคำสั่งซื้อนี้")
	case errors.Is(err, store.ErrOrderProcessed):
		writeError(w, http.StatusConflict, "ORDER_ALREADY_PROCESSED", "คำสั่งซื้อนี้ได้รับการตรวจสอบไปแล้ว กรุณารีเฟรชรายการ")
	default:
		internal(w, err)
	}
}

func (h *Handler) MyLibrary(w http.ResponseWriter, r *http.Request) {
	memberID := mw.WhoFrom(r.Context()).ID

	books, err := h.Store.MemberLibrary(r.Context(), memberID)
	if err != nil {
		internal(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": books})
}

// ใช้ตรวจสอบรูปแบบ ID ได้ในกรณีที่ต้องนำไปใช้เพิ่มเติม
