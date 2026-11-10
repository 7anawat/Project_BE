package store

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrOrderProcessed = errors.New("order already processed")
	ErrInvalidOrder   = errors.New("invalid order")
)

type OrderItem struct {
	ID       int64   `json:"id"`
	BookID   int64   `json:"bookId"`
	Title    string  `json:"title"`
	Author   string  `json:"author"`
	PriceTHB float64 `json:"priceTHB"`
	CoverURL *string `json:"coverUrl"`
	Quantity int     `json:"quantity"`
}

type Order struct {
	ID           int64       `json:"id"`
	MemberID     int64       `json:"memberId"`
	MemberName   string      `json:"memberName"`
	MemberEmail  string      `json:"memberEmail"`
	TotalAmount  float64     `json:"totalAmount"`
	TransferBank *string     `json:"transferBank"`
	SlipPath     string      `json:"-"`
	SlipURL      string      `json:"slipUrl,omitempty"`
	SlipNote     *string     `json:"slipNote"`
	Status       string      `json:"status"`
	ReviewedBy   *int64      `json:"reviewedBy"`
	ReviewedAt   *time.Time  `json:"reviewedAt"`
	RejectReason *string     `json:"rejectReason"`
	CreatedAt    time.Time   `json:"createdAt"`
	Items        []OrderItem `json:"items"`
}

// CreateOrder calculates the total from the database.
// The client only supplies book IDs, never trusted prices or totals.
func (s *Store) CreateOrder(
	ctx context.Context,
	memberID int64,
	bookIDs []int64,
	slipPath string,
	transferBank string,
	slipNote string,
) (Order, error) {
	if memberID <= 0 || len(bookIDs) == 0 || len(bookIDs) > 50 || slipPath == "" {
		return Order{}, ErrInvalidOrder
	}

	seen := make(map[int64]bool, len(bookIDs))
	for _, id := range bookIDs {
		if id <= 0 || seen[id] {
			return Order{}, ErrInvalidOrder
		}
		seen[id] = true
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback(ctx)

	items := make([]OrderItem, 0, len(bookIDs))
	var total float64

	for _, bookID := range bookIDs {
		var item OrderItem
		var price *float64

		err := tx.QueryRow(ctx, `
			SELECT id, title, author, price_thb, cover_url
			FROM books
			WHERE id = $1
		`, bookID).Scan(
			&item.BookID,
			&item.Title,
			&item.Author,
			&price,
			&item.CoverURL,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, ErrNotFound
		}
		if err != nil {
			return Order{}, err
		}
		if price == nil || *price < 0 {
			return Order{}, fmt.Errorf("book %d has no valid price", bookID)
		}

		item.PriceTHB = *price
		item.Quantity = 1
		total += *price
		items = append(items, item)
	}

	total = math.Round(total*100) / 100

	var order Order
	order.MemberID = memberID
	order.TotalAmount = total
	order.TransferBank = stringPointer(transferBank)
	order.SlipNote = stringPointer(slipNote)
	order.SlipPath = slipPath
	order.Status = "pending"
	order.Items = items

	err = tx.QueryRow(ctx, `
		INSERT INTO orders (
			member_id, total_amount, transfer_bank,
			slip_path, slip_note, status
		)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id, created_at
	`,
		memberID,
		total,
		transferBank,
		slipPath,
		nullableString(slipNote),
	).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return Order{}, err
	}

	for i := range items {
		item := &items[i]
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (
				order_id, book_id, title, author,
				price_thb, cover_url, quantity
			)
			VALUES ($1, $2, $3, $4, $5, $6, 1)
		`,
			order.ID,
			item.BookID,
			item.Title,
			item.Author,
			item.PriceTHB,
			item.CoverURL,
		)
		if err != nil {
			return Order{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Order{}, err
	}

	return order, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func (s *Store) ListMemberOrders(ctx context.Context, memberID int64) ([]Order, error) {
	return s.listOrders(ctx, `
		SELECT
			o.id, o.member_id, m.display_name, m.email,
			o.total_amount, o.transfer_bank, o.slip_path,
			o.slip_note, o.status, o.reviewed_by,
			o.reviewed_at, o.reject_reason, o.created_at
		FROM orders o
		JOIN members m ON m.id = o.member_id
		WHERE o.member_id = $1
		ORDER BY o.created_at DESC, o.id DESC
	`, memberID)
}

func (s *Store) ListAllOrders(ctx context.Context) ([]Order, error) {
	return s.listOrders(ctx, `
		SELECT
			o.id, o.member_id, m.display_name, m.email,
			o.total_amount, o.transfer_bank, o.slip_path,
			o.slip_note, o.status, o.reviewed_by,
			o.reviewed_at, o.reject_reason, o.created_at
		FROM orders o
		JOIN members m ON m.id = o.member_id
		ORDER BY
			CASE WHEN o.status = 'pending' THEN 0 ELSE 1 END,
			o.created_at DESC, o.id DESC
	`)
}

func (s *Store) listOrders(
	ctx context.Context,
	query string,
	args ...any,
) ([]Order, error) {
	rows, err := s.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	orders := make([]Order, 0)
	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID,
			&order.MemberID,
			&order.MemberName,
			&order.MemberEmail,
			&order.TotalAmount,
			&order.TransferBank,
			&order.SlipPath,
			&order.SlipNote,
			&order.Status,
			&order.ReviewedBy,
			&order.ReviewedAt,
			&order.RejectReason,
			&order.CreatedAt,
		)
		if err != nil {
			rows.Close()
			return nil, err
		}
		orders = append(orders, order)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}

	for i := range orders {
		items, err := s.listOrderItems(ctx, orders[i].ID)
		if err != nil {
			return nil, err
		}
		orders[i].Items = items
	}

	return orders, nil
}

func (s *Store) listOrderItems(ctx context.Context, orderID int64) ([]OrderItem, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, book_id, title, author, price_thb, cover_url, quantity
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]OrderItem, 0)
	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(
			&item.ID,
			&item.BookID,
			&item.Title,
			&item.Author,
			&item.PriceTHB,
			&item.CoverURL,
			&item.Quantity,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// ApproveOrder changes status and grants the books in one transaction.
// A second admin attempting the same approval receives ErrOrderProcessed.
func (s *Store) ApproveOrder(
	ctx context.Context,
	orderID int64,
	adminID int64,
) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var memberID int64
	err = tx.QueryRow(ctx, `
		UPDATE orders
		SET status = 'approved',
			reviewed_by = $2,
			reviewed_at = NOW(),
			reject_reason = NULL
		WHERE id = $1 AND status = 'pending'
		RETURNING member_id
	`, orderID, adminID).Scan(&memberID)

	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		checkErr := tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1)`,
			orderID,
		).Scan(&exists)
		if checkErr != nil {
			return 0, checkErr
		}
		if !exists {
			return 0, ErrNotFound
		}
		return 0, ErrOrderProcessed
	}
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO member_books (member_id, book_id, order_id)
		SELECT $1, book_id, $2
		FROM order_items
		WHERE order_id = $2
		ON CONFLICT (member_id, book_id) DO NOTHING
	`, memberID, orderID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return memberID, nil
}

// RejectOrder only rejects an order that is still pending.
func (s *Store) RejectOrder(
	ctx context.Context,
	orderID int64,
	adminID int64,
	reason string,
) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	tag, err := s.Pool.Exec(ctx, `
		UPDATE orders
		SET status = 'rejected',
			reviewed_by = $2,
			reviewed_at = NOW(),
			reject_reason = $3
		WHERE id = $1 AND status = 'pending'
	`, orderID, adminID, reason)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		var exists bool
		err := s.Pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1)`,
			orderID,
		).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNotFound
		}
		return ErrOrderProcessed
	}

	return nil
}

type LibraryBook struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Category    *string   `json:"category"`
	Description *string   `json:"description"`
	CoverURL    *string   `json:"coverUrl"`
	PurchasedAt time.Time `json:"purchasedAt"`
}

func (s *Store) MemberLibrary(ctx context.Context, memberID int64) ([]LibraryBook, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT
			b.id, b.title, b.author, b.category,
			b.description, b.cover_url, mb.purchased_at
		FROM member_books mb
		JOIN books b ON b.id = mb.book_id
		WHERE mb.member_id = $1
		ORDER BY mb.purchased_at DESC
	`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := make([]LibraryBook, 0)
	for rows.Next() {
		var book LibraryBook
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Author,
			&book.Category,
			&book.Description,
			&book.CoverURL,
			&book.PurchasedAt,
		); err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, rows.Err()
}
