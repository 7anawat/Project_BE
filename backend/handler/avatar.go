package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"bookhub/mw"
	"bookhub/store"
)

const maxAvatarSize int64 = 5 << 20 // 5 MB

// populateAvatarURL สร้าง Signed URL สำหรับรูปโปรไฟล์
func (h *Handler) populateAvatarURL(
	ctx context.Context,
	member *store.Member,
) error {
	member.AvatarURL = ""

	if member.AvatarPath == "" {
		return nil
	}

	if h.Storage == nil {
		return fmt.Errorf("supabase storage is not configured")
	}

	avatarURL, err := h.Storage.SignedURL(ctx, member.AvatarPath)
	if err != nil {
		return err
	}

	member.AvatarURL = avatarURL
	return nil
}

// UploadMyAvatar อัปโหลดรูปโปรไฟล์ของสมาชิกที่เข้าสู่ระบบ
func (h *Handler) UploadMyAvatar(w http.ResponseWriter, r *http.Request) {
	if h.Storage == nil {
		internal(w, fmt.Errorf("supabase storage is not configured"))
		return
	}

	// จำกัดขนาด request รวมเผื่อ multipart overhead
	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize+(1<<20))

	if err := r.ParseMultipartForm(maxAvatarSize + (1 << 20)); err != nil {
		var maxErr *http.MaxBytesError

		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge,
				"FILE_TOO_LARGE", "ไฟล์มีขนาดใหญ่เกินไป")
			return
		}

		writeError(w, http.StatusBadRequest,
			"BAD_MULTIPART", "ข้อมูลรูปภาพไม่ถูกต้อง")
		return
	}

	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		writeError(w, http.StatusBadRequest,
			"AVATAR_REQUIRED", "กรุณาเลือกรูปโปรไฟล์")
		return
	}
	defer file.Close()

	// อ่านไฟล์ไม่เกิน 5 MB + 1 byte เพื่อเช็กขนาดจริง
	data, err := io.ReadAll(io.LimitReader(file, maxAvatarSize+1))
	if err != nil {
		internal(w, err)
		return
	}

	if len(data) == 0 {
		writeError(w, http.StatusBadRequest,
			"EMPTY_FILE", "ไฟล์รูปภาพว่างเปล่า")
		return
	}

	if int64(len(data)) > maxAvatarSize {
		writeError(w, http.StatusRequestEntityTooLarge,
			"FILE_TOO_LARGE", "รูปโปรไฟล์ต้องมีขนาดไม่เกิน 5 MB")
		return
	}

	// ตรวจชนิดไฟล์จากเนื้อหาจริง ไม่เชื่อเฉพาะนามสกุลไฟล์
	contentType := http.DetectContentType(data)

	var extension string

	switch contentType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/webp":
		extension = ".webp"
	default:
		writeError(w, http.StatusBadRequest,
			"INVALID_IMAGE_TYPE", "รองรับเฉพาะ JPG, PNG และ WEBP")
		return
	}

	who := mw.WhoFrom(r.Context())

	// สร้างชื่อไฟล์สุ่ม ป้องกันชื่อซ้ำ
	var randomName [16]byte
	if _, err := rand.Read(randomName[:]); err != nil {
		internal(w, err)
		return
	}

	objectPath := fmt.Sprintf(
		"avatars/%d/%s%s",
		who.ID,
		hex.EncodeToString(randomName[:]),
		extension,
	)

	// อัปโหลดรูปใหม่ไป Supabase
	if err := h.Storage.Upload(
		r.Context(),
		objectPath,
		data,
		contentType,
	); err != nil {
		internal(w, err)
		return
	}

	// สร้าง URL ก่อนบันทึกฐานข้อมูล
	avatarURL, err := h.Storage.SignedURL(r.Context(), objectPath)
	if err != nil {
		if cleanupErr := h.Storage.Remove(r.Context(), objectPath); cleanupErr != nil {
			log.Printf("avatar cleanup failed: %v", cleanupErr)
		}
		internal(w, err)
		return
	}

	// บันทึก path ใหม่ และรับ path รูปเก่ากลับมา
	oldPath, err := h.Store.UpdateMemberAvatar(
		r.Context(),
		who.ID,
		objectPath,
	)

	if errors.Is(err, store.ErrNotFound) {
		if cleanupErr := h.Storage.Remove(r.Context(), objectPath); cleanupErr != nil {
			log.Printf("avatar cleanup failed: %v", cleanupErr)
		}

		writeError(w, http.StatusUnauthorized,
			"LOGIN_REQUIRED", "กรุณาเข้าสู่ระบบใหม่")
		return
	}

	if err != nil {
		if cleanupErr := h.Storage.Remove(r.Context(), objectPath); cleanupErr != nil {
			log.Printf("avatar cleanup failed: %v", cleanupErr)
		}

		internal(w, err)
		return
	}

	// ลบรูปเก่าหลังจากบันทึกรูปใหม่สำเร็จแล้ว
	// ลบเฉพาะ path ภายในโฟลเดอร์ของสมาชิกคนนี้
	expectedPrefix := fmt.Sprintf("avatars/%d/", who.ID)

	if oldPath != "" &&
		oldPath != objectPath &&
		strings.HasPrefix(oldPath, expectedPrefix) {

		if err := h.Storage.Remove(r.Context(), oldPath); err != nil {
			log.Printf("old avatar cleanup failed: %v", err)
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"avatarUrl": avatarURL,
	})
}

// ใช้ตรวจสอบว่าไฟล์นี้ยังอ้างอิง context และ store ถูกต้อง
var _ context.Context
