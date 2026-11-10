-- BookHub database schema (PostgreSQL 14+)
-- ร้านหนังสือออนไลน์: ข้อมูลหนังสือ, สมาชิก, รีวิว, wishlist

BEGIN;

-- ---------------------------------------------------------------------------
-- books: ข้อมูลหนังสือ 1 แถวต่อ 1 เล่ม
-- ---------------------------------------------------------------------------
CREATE TABLE books (
    id                 BIGSERIAL     PRIMARY KEY,
    title              VARCHAR(500)  NOT NULL,
    author             VARCHAR(255)  NOT NULL,
    category           VARCHAR(100),                -- หมวดหมู่ เช่น นวนิยาย, จิตวิทยา
    isbn               VARCHAR(20),
    publisher          VARCHAR(255),
    publication_year   INTEGER,
    description        TEXT,
    price_thb          NUMERIC(10,2),
    stock_quantity     INTEGER       NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    cover_url          VARCHAR(500),
    source_url         VARCHAR(500),
    language           CHAR(2)       NOT NULL DEFAULT 'th',
    created_at         TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_books_category ON books (category);
CREATE INDEX idx_books_title_lower ON books (lower(title));
CREATE INDEX idx_books_author_lower ON books (lower(author));

-- ---------------------------------------------------------------------------
-- members: สมาชิกของเว็บ
-- ---------------------------------------------------------------------------
CREATE TABLE members (
    id             BIGSERIAL     PRIMARY KEY,
    email          VARCHAR(255)  NOT NULL,
    password_hash  VARCHAR(255)  NOT NULL,
    display_name   VARCHAR(50)   NOT NULL CHECK (char_length(btrim(display_name)) >= 1),
    avatar_url     VARCHAR(500),
    role           VARCHAR(10)   NOT NULL DEFAULT 'member' CHECK (role IN ('member', 'admin')),
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_members_email ON members (lower(email));

-- ---------------------------------------------------------------------------
-- reviews: รีวิวหนังสือ (สมาชิก 1 คนรีวิวหนังสือ 1 เล่มได้ 1 ครั้ง)
-- ---------------------------------------------------------------------------
CREATE TABLE reviews (
    id          BIGSERIAL    PRIMARY KEY,
    member_id   BIGINT       NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    book_id     BIGINT       NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    text        TEXT         NOT NULL CHECK (char_length(btrim(text)) >= 10),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (member_id, book_id)
);

CREATE INDEX idx_reviews_book_created ON reviews (book_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- votes: คะแนนที่สมาชิกให้หนังสือ 1–10
-- ---------------------------------------------------------------------------
CREATE TABLE votes (
    member_id   BIGINT       NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    book_id     BIGINT       NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    score       SMALLINT     NOT NULL CHECK (score BETWEEN 1 AND 10),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    PRIMARY KEY (member_id, book_id)
);

CREATE INDEX idx_votes_book ON votes (book_id);

CREATE VIEW book_vote_stats AS
SELECT book_id,
       count(*)               AS vote_count,
       round(avg(score), 1)   AS vote_average
FROM votes
GROUP BY book_id;

-- ---------------------------------------------------------------------------
-- wishlist_items: หนังสือที่สมาชิกอยากอ่าน
-- ---------------------------------------------------------------------------
CREATE TABLE wishlist_items (
    member_id   BIGINT        NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    book_id     BIGINT        NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    note        VARCHAR(200),
    added_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
    PRIMARY KEY (member_id, book_id)
);

CREATE INDEX idx_wishlist_member_added ON wishlist_items (member_id, added_at DESC);

COMMIT;

