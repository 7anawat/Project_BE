package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrDuplicate = errors.New("duplicate")
	ErrNoBook    = errors.New("book not found")
)

type ReviewAuthor struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"displayName"`
}

type Review struct {
	ID        int64        `json:"id"`
	MemberID  int64        `json:"memberId"`
	BookID    int64        `json:"bookId"`
	Text      string       `json:"text"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	Member    ReviewAuthor `json:"member"`
	Score     *int16       `json:"score"`
}

func (s *Store) ListReviews(ctx context.Context, bookID int64) ([]Review, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT r.id, r.member_id, r.book_id, r.text, r.created_at, r.updated_at,
                m.display_name, v.score
         FROM reviews r
         JOIN members m ON m.id = r.member_id
         LEFT JOIN votes v ON v.member_id = r.member_id AND v.book_id = r.book_id
         WHERE r.book_id = $1
         ORDER BY r.created_at DESC`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Review{}
	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.ID, &r.MemberID, &r.BookID, &r.Text,
			&r.CreatedAt, &r.UpdatedAt, &r.Member.DisplayName, &r.Score); err != nil {
			return nil, err
		}
		r.Member.ID = r.MemberID
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) AddReview(ctx context.Context, memberID, bookID int64, text string) (Review, error) {
	r := Review{MemberID: memberID, BookID: bookID, Text: text}
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO reviews (member_id, book_id, text)
         VALUES ($1, $2, $3)
         RETURNING id, created_at, updated_at`, memberID, bookID, text).
		Scan(&r.ID, &r.CreatedAt, &r.UpdatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // ห้ามรีวิวซ้ำ
			return Review{}, ErrDuplicate
		case "23503": // หนังสือไม่มีจริง
			return Review{}, ErrNoBook
		}
	}
	return r, err
}

