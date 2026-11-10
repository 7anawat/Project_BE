package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"bookhub/auth"
	"bookhub/handler"
	"bookhub/mw"
	"bookhub/store"
)

func main() {
	// Load environment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Println("ไม่พบไฟล์ .env ใช้ Environment Variables จากระบบ")
	}
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET ต้องยาวอย่างน้อย 32 ตัวอักษร")
	}
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("ต่อฐานข้อมูลไม่ติด ", err)
	}

	tickets := &auth.Issuer{Secret: []byte(secret), TTL: time.Hour}
	h := &handler.Handler{Store: &store.Store{Pool: pool}, Tickets: tickets}
	login := func(f http.HandlerFunc) http.HandlerFunc { return mw.RequireLogin(tickets, f) }

	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("GET /api/me", login(h.Me))

	// Books (public)
	mux.HandleFunc("GET /api/books", h.SearchBooks)
	mux.HandleFunc("GET /api/books/{id}", h.GetBook)
	mux.HandleFunc("GET /api/books/{id}/reviews", h.ListReviews)

	// Reviews (require login)
	mux.HandleFunc("POST /api/books/{id}/reviews", login(h.AddReview))
	mux.HandleFunc("DELETE /api/reviews/{reviewId}", login(h.DeleteReview))

	// Categories (public)
	mux.HandleFunc("GET /api/categories", h.ListCategories)

	// Google Books API (admin only)
	mux.HandleFunc(
		"GET /api/admin/google-books",
		login(mw.RequireRole("admin", h.SearchGoogleBooks)),
	)
	// Admin (require login + admin role)
	mux.HandleFunc("GET /api/admin/members", login(mw.RequireRole("admin", h.AdminMembers)))
	mux.HandleFunc("PUT /api/admin/members/{memberId}/role", login(mw.RequireRole("admin", h.AdminSetRole)))
	mux.HandleFunc("GET /api/admin/stats", login(mw.RequireRole("admin", h.AdminStats)))
	mux.HandleFunc("POST /api/admin/books", login(mw.RequireRole("admin", h.AdminCreateBook)))
	mux.HandleFunc("GET /api/admin/books/{id}", login(mw.RequireRole("admin", h.AdminGetBook)))
	mux.HandleFunc("PUT /api/admin/books/{id}", login(mw.RequireRole("admin", h.AdminUpdateBook)))
	mux.HandleFunc("DELETE /api/admin/books/{id}", login(mw.RequireRole("admin", h.AdminDeleteBook)))

	// Wishlist (require login)
	mux.HandleFunc("GET /api/me/wishlist", login(h.GetWishlist))
	mux.HandleFunc("PUT /api/me/wishlist/{bookId}", login(h.AddToWishlist))
	mux.HandleFunc("DELETE /api/me/wishlist/{bookId}", login(h.RemoveFromWishlist))

	origins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := "0.0.0.0:" + port

	log.Println("NungsuKub API listening on", addr)
	log.Println("Allowed Origins:", origins)

	log.Fatal(http.ListenAndServe(addr, mw.Logging(mw.CORS(origins, mux))))
}
