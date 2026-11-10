-- BookHub: ลบหนังสือที่ไม่มี URL รูปปกเท่านั้น
-- หมายเหตุ: Schema ปัจจุบันตั้ง ON DELETE CASCADE ไว้กับ reviews, votes และ wishlist_items
-- การลบหนังสือเหล่านี้จะลบข้อมูลที่ผูกกับหนังสือเหล่านั้นตามไปด้วย

BEGIN;

DELETE FROM books
WHERE cover_url IS NULL OR btrim(cover_url) = ''
RETURNING id, title, isbn;

COMMIT;

SELECT COUNT(*) AS remaining_books FROM books;
SELECT COUNT(*) AS books_with_cover_url
FROM books
WHERE cover_url IS NOT NULL AND btrim(cover_url) <> '';
