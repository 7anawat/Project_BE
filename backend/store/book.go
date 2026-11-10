package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type Book struct {
	ID              int64    `json:"id"`
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
	SourceURL       *string  `json:"sourceUrl"`
	Language        string   `json:"language"`
}

// SearchBooks ค้นหาหนังสือจากชื่อหรือผู้แต่ง
func (s *Store) SearchBooks(ctx context.Context, q, category string) ([]Book, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, title, author, category, isbn, publisher, publication_year,
                description, price_thb, stock_quantity, cover_url, source_url, language
         FROM books
         WHERE (lower(title) LIKE '%' || lower($1) || '%'
             OR lower(author) LIKE '%' || lower($1) || '%')
           AND ($2 = '' OR category = $2)
         ORDER BY title ASC`, q, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Book{}
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Category, &b.ISBN,
			&b.Publisher, &b.PublicationYear, &b.Description, &b.PriceTHB,
			&b.StockQuantity, &b.CoverURL, &b.SourceURL, &b.Language); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// GetBook ดึงหนังสือเล่มเดียวตาม id
func (s *Store) GetBook(ctx context.Context, id int64) (Book, error) {
	var b Book
	err := s.Pool.QueryRow(ctx,
		`SELECT id, title, author, category, isbn, publisher, publication_year,
                description, price_thb, stock_quantity, cover_url, source_url, language
         FROM books WHERE id = $1`, id).
		Scan(&b.ID, &b.Title, &b.Author, &b.Category, &b.ISBN,
			&b.Publisher, &b.PublicationYear, &b.Description, &b.PriceTHB,
			&b.StockQuantity, &b.CoverURL, &b.SourceURL, &b.Language)
	if errors.Is(err, pgx.ErrNoRows) {
		return Book{}, ErrNotFound
	}
	return b, err
}

// ListCategories ดึงหมวดหมู่ทั้งหมดที่มีอยู่จริง
func (s *Store) ListCategories(ctx context.Context) ([]string, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT DISTINCT category FROM books WHERE category IS NOT NULL ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

type BookInput struct {
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

func (s *Store) CreateBook(ctx context.Context, in BookInput) (Book, error) {
	var b Book
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO books (title, author, category, isbn, publisher, publication_year,
                description, price_thb, stock_quantity, cover_url, language)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
         RETURNING id, title, author, category, isbn, publisher, publication_year,
                   description, price_thb, stock_quantity, cover_url, source_url, language`,
		in.Title, in.Author, in.Category, in.ISBN, in.Publisher, in.PublicationYear,
		in.Description, in.PriceTHB, in.StockQuantity, in.CoverURL, in.Language).
		Scan(&b.ID, &b.Title, &b.Author, &b.Category, &b.ISBN,
			&b.Publisher, &b.PublicationYear, &b.Description, &b.PriceTHB,
			&b.StockQuantity, &b.CoverURL, &b.SourceURL, &b.Language)
	return b, err
}

func (s *Store) UpdateBook(ctx context.Context, id int64, in BookInput) (Book, error) {
	var b Book
	err := s.Pool.QueryRow(ctx,
		`UPDATE books SET title=$1, author=$2, category=$3, isbn=$4, publisher=$5,
                publication_year=$6, description=$7, price_thb=$8, stock_quantity=$9,
                cover_url=$10, language=$11, updated_at=now()
         WHERE id = $12
         RETURNING id, title, author, category, isbn, publisher, publication_year,
                   description, price_thb, stock_quantity, cover_url, source_url, language`,
		in.Title, in.Author, in.Category, in.ISBN, in.Publisher, in.PublicationYear,
		in.Description, in.PriceTHB, in.StockQuantity, in.CoverURL, in.Language, id).
		Scan(&b.ID, &b.Title, &b.Author, &b.Category, &b.ISBN,
			&b.Publisher, &b.PublicationYear, &b.Description, &b.PriceTHB,
			&b.StockQuantity, &b.CoverURL, &b.SourceURL, &b.Language)
	if errors.Is(err, pgx.ErrNoRows) {
		return Book{}, ErrNotFound
	}
	return b, err
}

func (s *Store) DeleteBook(ctx context.Context, id int64) error {
	tag, err := s.Pool.Exec(ctx, `DELETE FROM books WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type AdminStats struct {
	Books        int64 `json:"books"`
	Members      int64 `json:"members"`
	Admins       int64 `json:"admins"`
	Reviews      int64 `json:"reviews"`
	WishlistItems int64 `json:"wishlistItems"`
}

func (s *Store) GetAdminStats(ctx context.Context) (AdminStats, error) {
	var st AdminStats
	err := s.Pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM books),
			(SELECT count(*) FROM members),
			(SELECT count(*) FROM members WHERE role='admin'),
			(SELECT count(*) FROM reviews),
			(SELECT count(*) FROM wishlist_items)
	`).Scan(&st.Books, &st.Members, &st.Admins, &st.Reviews, &st.WishlistItems)
	return st, err
}

