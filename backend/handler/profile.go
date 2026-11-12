package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"bookhub/auth"
	"bookhub/mw"
	"bookhub/store"
)

type updateProfileReq struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type changePasswordReq struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// UpdateMyProfile แก้ไขข้อมูลสมาชิกที่ Login อยู่
func (h *Handler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	var req updateProfileReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "ข้อมูล JSON ไม่ถูกต้อง")
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.DisplayName)

	switch {
	case !strings.Contains(email, "@") || len(email) > 255:
		writeError(w, http.StatusBadRequest, "BAD_EMAIL", "อีเมลไม่ถูกต้อง")
		return

	case name == "" || utf8.RuneCountInString(name) > 50:
		writeError(w, http.StatusBadRequest, "BAD_NAME", "ชื่อต้องมี 1-50 ตัวอักษร")
		return
	}

	who := mw.WhoFrom(r.Context())

	member, err := h.Store.UpdateMemberProfile(
		r.Context(),
		who.ID,
		email,
		name,
	)

	switch {
	case errors.Is(err, store.ErrEmailTaken):
		writeError(w, http.StatusConflict, "EMAIL_TAKEN", "อีเมลนี้ถูกใช้งานแล้ว")
		return

	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusUnauthorized, "LOGIN_REQUIRED", "กรุณาเข้าสู่ระบบใหม่")
		return

	case err != nil:
		internal(w, err)
		return
	}
	if err := h.populateAvatarURL(r.Context(), &member); err != nil {
		internal(w, err)
		return
	}

	writeJSON(w, http.StatusOK, member)
}

// ChangeMyPassword ตรวจรหัสเดิมและบันทึกรหัสใหม่ด้วย bcrypt
func (h *Handler) ChangeMyPassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "ข้อมูล JSON ไม่ถูกต้อง")
		return
	}

	if req.CurrentPassword == "" {
		writeError(w, http.StatusBadRequest, "CURRENT_PASSWORD_REQUIRED", "กรุณากรอกรหัสผ่านเดิม")
		return
	}

	if utf8.RuneCountInString(req.NewPassword) < 8 || len(req.NewPassword) > 72 {
		writeError(w, http.StatusBadRequest, "BAD_PASSWORD", "รหัสผ่านใหม่ต้องมีอย่างน้อย 8 ตัวอักษร และไม่เกิน 72 ไบต์")
		return
	}

	if req.CurrentPassword == req.NewPassword {
		writeError(w, http.StatusBadRequest, "PASSWORD_UNCHANGED", "รหัสผ่านใหม่ต้องไม่ซ้ำกับรหัสเดิม")
		return
	}

	who := mw.WhoFrom(r.Context())

	// 1. ดึง Password Hash เดิม
	oldHash, err := h.Store.MemberPasswordHashByID(
		r.Context(),
		who.ID,
	)

	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "LOGIN_REQUIRED", "กรุณาเข้าสู่ระบบใหม่")
		return
	}

	if err != nil {
		internal(w, err)
		return
	}

	// 2. ตรวจสอบรหัสผ่านเดิม
	if !auth.CheckPassword(oldHash, req.CurrentPassword) {
		writeError(w, http.StatusBadRequest, "WRONG_CURRENT_PASSWORD", "รหัสผ่านเดิมไม่ถูกต้อง")
		return
	}

	// 3. Hash รหัสผ่านใหม่ด้วย bcrypt
	newHash, err := auth.HashPassword(req.NewPassword)

	if err != nil {
		internal(w, err)
		return
	}

	// 4. บันทึก Hash ใหม่ลง Database
	err = h.Store.UpdateMemberPasswordHash(
		r.Context(),
		who.ID,
		newHash,
	)

	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "LOGIN_REQUIRED", "กรุณาเข้าสู่ระบบใหม่")
		return
	}

	if err != nil {
		internal(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "เปลี่ยนรหัสผ่านเรียบร้อยแล้ว",
	})
}
