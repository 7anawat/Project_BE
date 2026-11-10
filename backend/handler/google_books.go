package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// โครงสร้างข้อมูลที่ได้รับจาก Google Books
type googleBooksResponse struct {
	TotalItems int                 `json:"totalItems"`
	Items      []googleBooksVolume `json:"items"`
}

type googleBooksVolume struct {
	ID         string          `json:"id"`
	VolumeInfo googleBooksInfo `json:"volumeInfo"`
}

type googleBooksInfo struct {
	Title         string   `json:"title"`
	Authors       []string `json:"authors"`
	Publisher     string   `json:"publisher"`
	PublishedDate string   `json:"publishedDate"`
	Description   string   `json:"description"`
	PageCount     int      `json:"pageCount"`
	Categories    []string `json:"categories"`
	Language      string   `json:"language"`

	IndustryIdentifiers []struct {
		Type       string `json:"type"`
		Identifier string `json:"identifier"`
	} `json:"industryIdentifiers"`

	ImageLinks struct {
		Thumbnail      string `json:"thumbnail"`
		SmallThumbnail string `json:"smallThumbnail"`
	} `json:"imageLinks"`
}

// ข้อมูลหนังสือที่ส่งกลับไปยัง Frontend
type googleBookResult struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Authors       []string `json:"authors"`
	Publisher     string   `json:"publisher"`
	PublishedDate string   `json:"publishedDate"`
	Description   string   `json:"description"`
	ISBN          string   `json:"isbn"`
	PageCount     int      `json:"pageCount"`
	Categories    []string `json:"categories"`
	CoverURL      string   `json:"coverUrl"`
	Language      string   `json:"language"`
}

// SearchGoogleBooks ค้นหาหนังสือจาก Google Books API
func (h *Handler) SearchGoogleBooks(
	w http.ResponseWriter,
	r *http.Request,
) {
	// รับคำค้นหา
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	if q == "" {
		writeError(w, http.StatusBadRequest,
			"EMPTY_QUERY", "กรุณากรอกชื่อหนังสือ")
		return
	}

	// อ่าน API Key จาก Environment Variable
	apiKey := os.Getenv("GOOGLE_BOOKS_API_KEY")

	if apiKey == "" {
		writeError(w, http.StatusInternalServerError,
			"MISSING_API_KEY", "ยังไม่ได้ตั้งค่า Google Books API Key")
		return
	}

	// กำหนดจำนวนผลลัพธ์
	maxResults := 10

	if n, err := strconv.Atoi(r.URL.Query().Get("maxResults")); err == nil {
		if n >= 1 && n <= 40 {
			maxResults = n
		}
	}

	// สร้าง URL สำหรับ Google Books API
	params := url.Values{}
	params.Set("q", q)
	params.Set("maxResults", strconv.Itoa(maxResults))
	params.Set("key", apiKey)

	apiURL := "https://www.googleapis.com/books/v1/volumes?" + params.Encode()

	req, err := http.NewRequestWithContext(
		r.Context(),
		http.MethodGet,
		apiURL,
		nil,
	)

	if err != nil {
		writeError(w, http.StatusInternalServerError,
			"REQUEST_ERROR", "สร้างคำขอไม่สำเร็จ")
		return
	}

	// เรียก Google Books API
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)

	if err != nil {
		writeError(w, http.StatusBadGateway,
			"GOOGLE_API_ERROR", "ไม่สามารถเชื่อมต่อ Google Books ได้")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		writeError(w, http.StatusBadGateway,
			"GOOGLE_API_ERROR", "Google Books API ตอบกลับผิดพลาด")
		return
	}

	// อ่านข้อมูล JSON
	var data googleBooksResponse

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		writeError(w, http.StatusBadGateway,
			"INVALID_RESPONSE", "อ่านข้อมูลจาก Google Books ไม่สำเร็จ")
		return
	}

	// แปลงข้อมูลให้ Frontend ใช้งานง่าย
	items := make([]googleBookResult, 0, len(data.Items))

	for _, book := range data.Items {
		info := book.VolumeInfo

		isbn := ""

		for _, identifier := range info.IndustryIdentifiers {
			if identifier.Type == "ISBN_13" {
				isbn = identifier.Identifier
				break
			}

			if identifier.Type == "ISBN_10" && isbn == "" {
				isbn = identifier.Identifier
			}
		}

		coverURL := info.ImageLinks.Thumbnail

		if coverURL == "" {
			coverURL = info.ImageLinks.SmallThumbnail
		}

		coverURL = strings.Replace(
			coverURL, "http://", "https://", 1,
		)

		items = append(items, googleBookResult{
			ID:            book.ID,
			Title:         info.Title,
			Authors:       info.Authors,
			Publisher:     info.Publisher,
			PublishedDate: info.PublishedDate,
			Description:   info.Description,
			ISBN:          isbn,
			PageCount:     info.PageCount,
			Categories:    info.Categories,
			CoverURL:      coverURL,
			Language:      info.Language,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"totalItems": data.TotalItems,
		"items":      items,
	})
}
