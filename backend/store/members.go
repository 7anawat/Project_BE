package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrEmailTaken = errors.New("email taken")

type Member struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	DisplayName  string `json:"displayName"`
	Role         string `json:"role"`
	CreatedAt    string `json:"createdAt,omitempty"`
	AvatarPath   string `json:"-"`
	AvatarURL    string `json:"avatarUrl,omitempty"`
	PasswordHash string `json:"-"`
}

func (s *Store) CreateMember(ctx context.Context, email, hash, name string) (Member, error) {
	m := Member{Email: email, DisplayName: name}
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO members (email, password_hash, display_name)
         VALUES ($1, $2, $3)
         RETURNING id, role`, email, hash, name).Scan(&m.ID, &m.Role)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Member{}, ErrEmailTaken
	}
	return m, err
}

func (s *Store) MemberByEmail(ctx context.Context, email string) (Member, error) {
	var m Member

	err := s.Pool.QueryRow(ctx,
		`SELECT id, email, display_name, role, password_hash, COALESCE(avatar_url, '')
         FROM members WHERE lower(email) = $1`, email).
		Scan(
			&m.ID,
			&m.Email,
			&m.DisplayName,
			&m.Role,
			&m.PasswordHash,
			&m.AvatarPath,
		)

	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}

	return m, err
}

func (s *Store) MemberByID(ctx context.Context, id int64) (Member, error) {
	var m Member

	err := s.Pool.QueryRow(ctx,
		`SELECT id, email, display_name, role,
               created_at::text, COALESCE(avatar_url, '')
         FROM members WHERE id = $1`, id).
		Scan(
			&m.ID,
			&m.Email,
			&m.DisplayName,
			&m.Role,
			&m.CreatedAt,
			&m.AvatarPath,
		)

	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}

	return m, err
}

func (s *Store) UpdateMemberProfile(ctx context.Context, memberID int64, email, displayName string) (Member, error) {
	var m Member

	err := s.Pool.QueryRow(ctx,
		`UPDATE members
         SET email = $1, display_name = $2, updated_at = now()
         WHERE id = $3
         RETURNING id, email, display_name, role,
created_at::text, COALESCE(avatar_url, '')`,
		email, displayName, memberID,
	).Scan(
		&m.ID,
		&m.Email,
		&m.DisplayName,
		&m.Role,
		&m.CreatedAt,
		&m.AvatarPath,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Member{}, ErrEmailTaken
	}

	return m, err
}

func (s *Store) UpdateMemberAvatar(ctx context.Context, memberID int64, newAvatarPath string) (string, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	// ดึง Path ของรูปเดิม และล็อกแถวระหว่างทำ Transaction
	var oldAvatarPath string

	err = tx.QueryRow(ctx,
		`SELECT COALESCE(avatar_url, '')
         FROM members WHERE id = $1
         FOR UPDATE`,
		memberID,
	).Scan(&oldAvatarPath)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	// บันทึก Path ของรูปใหม่
	_, err = tx.Exec(ctx,
		`UPDATE members
         SET avatar_url = $1, updated_at = now()
         WHERE id = $2`,
		newAvatarPath, memberID,
	)

	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	// ส่ง Path รูปเดิมกลับไปให้ Handler จัดการลบภายหลัง
	return oldAvatarPath, nil
}

// ดึง Password Hash ปัจจุบันของสมาชิก
func (s *Store) MemberPasswordHashByID(ctx context.Context, memberID int64) (string, error) {
	var hash string

	err := s.Pool.QueryRow(ctx,
		`SELECT password_hash FROM members WHERE id = $1`,
		memberID,
	).Scan(&hash)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}

	return hash, err
}

// บันทึก Password Hash ใหม่
func (s *Store) UpdateMemberPasswordHash(ctx context.Context, memberID int64, passwordHash string) error {
	tag, err := s.Pool.Exec(ctx,
		`UPDATE members
         SET password_hash = $1, updated_at = now()
         WHERE id = $2`,
		passwordHash, memberID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *Store) ListMembers(ctx context.Context) ([]Member, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT id, email, display_name, role, created_at::text FROM members ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Member{}
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) SetMemberRole(ctx context.Context, memberID int64, role string) (Member, error) {
	var m Member
	err := s.Pool.QueryRow(ctx,
		`UPDATE members SET role = $1, updated_at = now()
         WHERE id = $2
         RETURNING id, email, display_name, role`, role, memberID).
		Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}
	return m, err
}

func (s *Store) DeleteReview(ctx context.Context, reviewID, memberID int64, isAdmin bool) error {
	tag, err := s.Pool.Exec(ctx,
		`DELETE FROM reviews
         WHERE id = $1 AND (member_id = $2 OR $3)`, reviewID, memberID, isAdmin)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
