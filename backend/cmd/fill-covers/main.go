package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5/pgxpool"
)

type googleBooksResponse struct {
	Items []struct {
		VolumeInfo struct {
			Title               string   `json:"title"`
			Authors             []string `json:"authors"`
			IndustryIdentifiers []struct {
				Type       string `json:"type"`
				Identifier string `json:"identifier"`
			} `json:"industryIdentifiers"`
			ImageLinks struct {
				Thumbnail      string `json:"thumbnail"`
				SmallThumbnail string `json:"smallThumbnail"`
			} `json:"imageLinks"`
		} `json:"volumeInfo"`
	} `json:"items"`
}

type openLibrarySearchResponse struct {
	Docs []struct {
		Title      string   `json:"title"`
		AuthorName []string `json:"author_name"`
		CoverID    int64    `json:"cover_i"`
		ISBN       []string `json:"isbn"`
	} `json:"docs"`
}

type bookRow struct {
	ID     int64
	ISBN   string
	Title  string
	Author string
}

type coverMatch struct {
	URL    string
	Method string
	Score  int
}

var parenPattern = regexp.MustCompile(`\(([^()]*)\)`)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func run() error {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	apiKey := strings.TrimSpace(os.Getenv("GOOGLE_BOOKS_API_KEY"))
	if databaseURL == "" {
		return fmt.Errorf("ไม่พบ DATABASE_URL: กรุณาตั้งค่า DATABASE_URL ใน PowerShell ก่อน")
	}
	if apiKey == "" {
		return fmt.Errorf("ไม่พบ GOOGLE_BOOKS_API_KEY: กรุณาตั้งค่า API Key ใน PowerShell ก่อน")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("เชื่อมต่อ PostgreSQL ไม่สำเร็จ: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ติดต่อ PostgreSQL ไม่สำเร็จ: %w", err)
	}

	rows, err := pool.Query(ctx, `
		SELECT id, COALESCE(isbn, ''), title, COALESCE(author, '')
		FROM books
		WHERE cover_url IS NULL OR btrim(cover_url) = ''
		ORDER BY id
	`)
	if err != nil {
		return fmt.Errorf("อ่านรายการหนังสือไม่สำเร็จ: %w", err)
	}
	books := make([]bookRow, 0)
	for rows.Next() {
		var b bookRow
		if err := rows.Scan(&b.ID, &b.ISBN, &b.Title, &b.Author); err != nil {
			rows.Close()
			return fmt.Errorf("อ่านข้อมูลหนังสือไม่สำเร็จ: %w", err)
		}
		books = append(books, b)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	if len(books) == 0 {
		fmt.Println("หนังสือทุกเล่มมี cover_url แล้ว")
		return nil
	}

	client := &http.Client{Timeout: 20 * time.Second}
	updated, byISBN, byTitle, byOpenLibrary, notFound, failed := 0, 0, 0, 0, 0, 0
	fmt.Printf("พบหนังสือ %d เล่มที่ยังไม่มีรูปปก\n", len(books))
	fmt.Println("ค้นหาจาก Google Books ก่อน หากไม่พบจะลอง Open Library ด้วย ISBN และชื่อ/ผู้แต่ง")
	fmt.Println("หมายเหตุ: จะบันทึกเฉพาะผลที่ชื่อหนังสือตรงพอสมควร เพื่อลดโอกาสใช้ปกผิดเล่ม")

	for i, b := range books {
		match, googleErr := findCover(ctx, client, apiKey, b)
		if match.URL == "" {
			olMatch, olErr := findOpenLibraryCover(ctx, client, b)
			if olErr == nil && olMatch.URL != "" {
				match = olMatch
			} else if googleErr != nil && olErr != nil {
				failed++
				fmt.Printf("[%d/%d] ID %d: ค้นหาจากทั้ง Google Books และ Open Library ไม่สำเร็จ\n", i+1, len(books), b.ID)
				continue
			}
		}
		if match.URL == "" {
			notFound++
			fmt.Printf("[%d/%d] ID %d: ยังไม่พบปกที่ตรงกับชื่อ/ISBN\n", i+1, len(books), b.ID)
			continue
		}

		_, err = pool.Exec(ctx,
			`UPDATE books SET cover_url = $1, updated_at = now() WHERE id = $2 AND (cover_url IS NULL OR btrim(cover_url) = '')`,
			match.URL, b.ID,
		)
		if err != nil {
			failed++
			fmt.Printf("[%d/%d] ID %d: บันทึก URL ไม่สำเร็จ: %v\n", i+1, len(books), b.ID, err)
			continue
		}
		updated++
		switch {
		case strings.HasPrefix(match.Method, "Open Library"):
			byOpenLibrary++
		case match.Method == "ISBN":
			byISBN++
		default:
			byTitle++
		}
		fmt.Printf("[%d/%d] เพิ่มรูปปกแล้ว (ค้นจาก %s): %s\n", i+1, len(books), match.Method, b.Title)
		time.Sleep(250 * time.Millisecond)
	}

	fmt.Println("\nสรุปผล")
	fmt.Printf("เพิ่มรูปปกสำเร็จ: %d เล่ม\n", updated)
	fmt.Printf("  - พบจาก ISBN: %d เล่ม\n", byISBN)
	fmt.Printf("  - พบจากชื่อ/ผู้แต่ง: %d เล่ม\n", byTitle)
	fmt.Printf("  - พบจาก Open Library: %d เล่ม\n", byOpenLibrary)
	fmt.Printf("ยังไม่พบรูปปก: %d เล่ม\n", notFound)
	fmt.Printf("เกิดข้อผิดพลาด: %d เล่ม\n", failed)
	fmt.Println("ตรวจสอบด้วย SQL: SELECT id, title, cover_url FROM books WHERE cover_url IS NOT NULL AND btrim(cover_url) <> '' ORDER BY id;")
	return nil
}

func findCover(ctx context.Context, client *http.Client, apiKey string, b bookRow) (coverMatch, error) {
	queries := make([]struct{ Query, Method string }, 0, 6)
	addQuery := func(q, method string) {
		q = strings.TrimSpace(q)
		if q == "" {
			return
		}
		for _, old := range queries {
			if old.Query == q {
				return
			}
		}
		queries = append(queries, struct{ Query, Method string }{q, method})
	}

	if strings.TrimSpace(b.ISBN) != "" {
		addQuery("isbn:"+strings.TrimSpace(b.ISBN), "ISBN")
	}

	variants := titleVariants(b.Title)
	for _, title := range variants {
		quoted := `"` + title + `"`
		addQuery("intitle:"+quoted, "TITLE")
		if strings.TrimSpace(b.Author) != "" {
			addQuery("intitle:"+quoted+" inauthor:\""+strings.TrimSpace(b.Author)+"\"", "TITLE_AUTHOR")
		}
	}
	// General full-text search as a final fallback.
	if len(variants) > 0 {
		addQuery(`"`+variants[0]+`"`, "TITLE")
	}

	var lastErr error
	best := coverMatch{}
	for _, q := range queries {
		data, err := searchGoogleBooks(ctx, client, apiKey, q.Query)
		if err != nil {
			lastErr = err
			continue
		}
		for _, item := range data.Items {
			info := item.VolumeInfo
			cover := info.ImageLinks.Thumbnail
			if cover == "" {
				cover = info.ImageLinks.SmallThumbnail
			}
			if cover == "" {
				continue
			}
			cover = strings.Replace(cover, "http://", "https://", 1)

			score := titleMatchScore(b.Title, info.Title)
			if hasISBN(info.IndustryIdentifiers, b.ISBN) {
				score = 200
			}
			if score > best.Score {
				method := q.Method
				if score == 200 {
					method = "ISBN"
				} else if method == "TITLE_AUTHOR" {
					method = "ชื่อ/ผู้แต่ง"
				} else {
					method = "ชื่อหนังสือ"
				}
				best = coverMatch{URL: cover, Method: method, Score: score}
			}
		}
		if best.Score == 200 {
			return best, nil
		}
	}
	if best.Score >= 70 {
		return best, nil
	}
	if lastErr != nil && len(queries) > 0 {
		return coverMatch{}, lastErr
	}
	return coverMatch{}, nil
}

func searchGoogleBooks(ctx context.Context, client *http.Client, apiKey, query string) (googleBooksResponse, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("maxResults", "10")
	params.Set("key", apiKey)
	endpoint := "https://www.googleapis.com/books/v1/volumes?" + params.Encode()

	reqCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return googleBooksResponse{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return googleBooksResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return googleBooksResponse{}, fmt.Errorf("Google Books HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var data googleBooksResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return googleBooksResponse{}, err
	}
	return data, nil
}

// findOpenLibraryCover is a one-time fallback for books not found in Google Books.
// Open Library asks clients to keep request rates low; this helper waits at least 1.1s
// between API calls. Existing cover_url values are never overwritten by the caller.
func findOpenLibraryCover(ctx context.Context, client *http.Client, b bookRow) (coverMatch, error) {
	if isbn := normalizeISBN(b.ISBN); isbn != "" {
		if ok, err := openLibraryCoverExists(ctx, client, isbn); err != nil {
			return coverMatch{}, err
		} else if ok {
			return coverMatch{
				URL:    "https://covers.openlibrary.org/b/isbn/" + url.PathEscape(isbn) + "-M.jpg?default=false",
				Method: "Open Library ISBN",
				Score:  200,
			}, nil
		}
	}

	// Search by each title variant (English text in parentheses is included).
	for _, title := range titleVariants(b.Title) {
		if strings.TrimSpace(title) == "" {
			continue
		}
		if err := waitOpenLibrary(); err != nil {
			return coverMatch{}, err
		}
		params := url.Values{}
		params.Set("title", title)
		params.Set("limit", "10")
		params.Set("fields", "title,author_name,cover_i,isbn")
		if strings.TrimSpace(b.Author) != "" {
			params.Set("author", b.Author)
		}
		endpoint := "https://openlibrary.org/search.json?" + params.Encode()
		data, err := getOpenLibrarySearch(ctx, client, endpoint)
		if err != nil {
			return coverMatch{}, err
		}

		best := coverMatch{}
		for _, doc := range data.Docs {
			if doc.CoverID <= 0 {
				continue
			}
			score := titleMatchScore(b.Title, doc.Title)
			if containsISBN(doc.ISBN, b.ISBN) {
				return coverMatch{
					URL:    fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg?default=false", doc.CoverID),
					Method: "Open Library ISBN/search",
					Score:  200,
				}, nil
			}
			// Require a strong title match to avoid attaching the wrong cover.
			if score >= 95 && score > best.Score {
				best = coverMatch{
					URL:    fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg?default=false", doc.CoverID),
					Method: "Open Library title",
					Score:  score,
				}
			}
		}
		if best.URL != "" {
			return best, nil
		}
	}
	return coverMatch{}, nil
}

func openLibraryCoverExists(ctx context.Context, client *http.Client, isbn string) (bool, error) {
	if err := waitOpenLibrary(); err != nil {
		return false, err
	}
	endpoint := "https://covers.openlibrary.org/b/isbn/" + url.PathEscape(isbn) + "-M.jpg?default=false"
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "BookHubLocalCoverLookup/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 512))
		return false, nil
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "image/") {
		return false, nil
	}
	// Read only a small prefix; the image itself is served by Open Library in the browser.
	buf := make([]byte, 32)
	n, _ := io.ReadFull(resp.Body, buf)
	return n > 8, nil
}

func getOpenLibrarySearch(ctx context.Context, client *http.Client, endpoint string) (openLibrarySearchResponse, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return openLibrarySearchResponse{}, err
	}
	req.Header.Set("User-Agent", "BookHubLocalCoverLookup/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return openLibrarySearchResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return openLibrarySearchResponse{}, fmt.Errorf("Open Library HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var data openLibrarySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return openLibrarySearchResponse{}, err
	}
	return data, nil
}

func waitOpenLibrary() error {
	time.Sleep(1100 * time.Millisecond)
	return nil
}

func containsISBN(candidates []string, isbn string) bool {
	want := normalizeISBN(isbn)
	if want == "" {
		return false
	}
	for _, candidate := range candidates {
		if normalizeISBN(candidate) == want {
			return true
		}
	}
	return false
}

func titleVariants(title string) []string {
	variants := []string{}
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		for _, old := range variants {
			if normalize(old) == normalize(s) {
				return
			}
		}
		variants = append(variants, s)
	}
	add(title)
	for _, m := range parenPattern.FindAllStringSubmatch(title, -1) {
		if len(m) > 1 {
			add(m[1])
		}
	}
	return variants
}

func titleMatchScore(localTitle, candidateTitle string) int {
	candidate := normalize(candidateTitle)
	if candidate == "" {
		return 0
	}
	best := 0
	for _, variant := range titleVariants(localTitle) {
		want := normalize(variant)
		if want == "" {
			continue
		}
		if want == candidate {
			return 120
		}
		if strings.Contains(candidate, want) || strings.Contains(want, candidate) {
			shorter, longer := len([]rune(want)), len([]rune(candidate))
			if shorter > 0 && float64(shorter)/float64(longer) >= 0.65 && best < 95 {
				best = 95
			}
		}
		// For titles with extra subtitles, require a high token overlap.
		wantTokens := strings.Fields(want)
		candidateTokens := strings.Fields(candidate)
		if len(wantTokens) > 1 && len(candidateTokens) > 1 {
			seen := map[string]bool{}
			for _, t := range candidateTokens {
				seen[t] = true
			}
			matched := 0
			for _, t := range wantTokens {
				if seen[t] {
					matched++
				}
			}
			ratio := float64(matched) / float64(len(wantTokens))
			if ratio >= 0.8 && best < 75 {
				best = 75
			}
		}
	}
	return best
}

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		} else if unicode.IsSpace(r) {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func hasISBN(ids []struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}, isbn string) bool {
	want := normalizeISBN(isbn)
	if want == "" {
		return false
	}
	for _, id := range ids {
		if normalizeISBN(id.Identifier) == want {
			return true
		}
	}
	return false
}

func normalizeISBN(s string) string {
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}
