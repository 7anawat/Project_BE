package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type WishlistItem struct {
	BookID        int64    `json:"bookId"`
	Title         string   `json:"title"`
	Author        string   `json:"author"`
	Category      *string  `json:"category"`
	PriceTHB      *float64 `json:"priceTHB"`
	CoverURL      *string  `json:"coverUrl"`
	StockQuantity int      `json:"stockQuantity"`
	Note          *string  `json:"note"`
	AddedAt       time.Time `json:"addedAt"`
}

func (s *Store) GetWishlist(ctx context.Context, memberID int64) ([]WishlistItem, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT w.book_id, b.title, b.author, b.category, b.price_thb, b.cover_url, b.stock_quantity, w.note, w.added_at
         FROM wishlist_items w
         JOIN books b ON b.id = w.book_id
         WHERE w.member_id = $1
         ORDER BY w.added_at DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []WishlistItem{}
	for rows.Next() {
		var item WishlistItem
		if err := rows.Scan(&item.BookID, &item.Title, &item.Author, &item.Category,
			&item.PriceTHB, &item.CoverURL, &item.StockQuantity, &item.Note, &item.AddedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

var ErrAlreadyInWishlist = errors.New("already in wishlist")

func (s *Store) AddToWishlist(ctx context.Context, memberID, bookID int64) error {
	_, err := s.Pool.Exec(ctx,
		`INSERT INTO wishlist_items (member_id, book_id) VALUES ($1, $2)`, memberID, bookID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrAlreadyInWishlist
	}
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrNoBook
	}
	return err
}

func (s *Store) RemoveFromWishlist(ctx context.Context, memberID, bookID int64) error {
	tag, err := s.Pool.Exec(ctx,
		`DELETE FROM wishlist_items WHERE member_id = $1 AND book_id = $2`, memberID, bookID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) IsInWishlist(ctx context.Context, memberID, bookID int64) (bool, error) {
	var exists bool
	err := s.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM wishlist_items WHERE member_id=$1 AND book_id=$2)`,
		memberID, bookID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return exists, err
}

