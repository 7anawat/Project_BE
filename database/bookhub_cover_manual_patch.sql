-- BookHub cover patch (verified title/ISBN matches)
-- Updates only rows that still have no cover_url. Existing covers are preserved.
-- Source images are hosted by external book retailers/publishers; for production,
-- prefer storing licensed cover files in your own storage/CDN.

BEGIN;

-- ID 3 | Money 101 | ISBN 9786160834563
UPDATE books SET cover_url = 'https://storage.naiin.com/system/application/bookstore/resource/product/201910/491276/1000225492_front_XXL.jpg?v=1725795565'
WHERE id = 3 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 6 | เดนดาว NEVER DIE | ISBN 9786168166956
UPDATE books SET cover_url = 'https://api.chulabook.com/images/pid-231417.JPG'
WHERE id = 6 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 7 | 52 เฮิรตซ์...คลื่นเสียงที่ไม่มีใครได้ยิน | ISBN 9786161846985
UPDATE books SET cover_url = 'https://storage.naiin.com/system/application/bookstore/resource/product/202202/541412/1000246409_front_XXL.jpg?v=1725894189'
WHERE id = 7 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 9 | ไนปุรานเริกส์ | ISBN 9786168296578
UPDATE books SET cover_url = 'https://storage.naiin.com/system/application/bookstore/resource/product/202604/701156/1000290052_front_XXL.jpg?v=1775125814'
WHERE id = 9 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 10 | Longevity by Design | ISBN 9786169442172
UPDATE books SET cover_url = 'https://www.asiabooks.com/media/catalog/product/cache/a5ac216be58c0cbce1cb04612ece96dc/9/7/9786169442172.jpg'
WHERE id = 10 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 11 | วิถีแห่งการลงทุนของผม บทเรียนจากเซียนหุ้นญี่ปุ่น | ISBN 9786169440031
UPDATE books SET cover_url = 'https://down-th.img.susercontent.com/file/th-11134207-7rash-m4stjsr2gud2e1'
WHERE id = 11 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 13 | คำสอนจากมหาเศรษฐีแห่งบาบิโลน | ISBN 9786168335819
UPDATE books SET cover_url = 'https://storage.naiin.com/system/application/bookstore/resource/product/202501/634674/1000278786_front_XXL.jpg?v=1736410577'
WHERE id = 13 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 16 | จัดการคนรับมือยาก 9 ไทป์ยังไงให้อยู่หมัด | ISBN 9786160474141
UPDATE books SET cover_url = 'https://platform-api.nanmeebooks.com/uploads/images/image-1769160429894.png'
WHERE id = 16 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 18 | Confidence: 8 ขั้นตอนสร้างความมั่นใจที่แท้จริง | ISBN 9786161885434
UPDATE books SET cover_url = 'https://storage.naiin.com/system/application/bookstore/resource/product/202512/691815/1000287475_front_XXL.jpg?v=1765420328'
WHERE id = 18 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 19 | จิตวิทยาแค่ 1% ทำให้คุณเหนือคน | ISBN 9786163651365
UPDATE books SET cover_url = 'https://down-th.img.susercontent.com/file/36084ebc3638ea10d9c9e5404d73da82'
WHERE id = 19 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 21 | คุณในอนาคต ต้องขอบคุณตัวเอง ที่อดทนสู้ในวันนี้ | ISBN 9789749760970
UPDATE books SET cover_url = 'https://www.asiabooks.com/media/catalog/product/cache/a5ac216be58c0cbce1cb04612ece96dc/9/7/9789749760970_1.png'
WHERE id = 21 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 32 | การเงิน 101 | ISBN 9786160837564
UPDATE books SET cover_url = 'https://down-th.img.susercontent.com/file/th-11134207-7qul3-lf0wlh9nep0157'
WHERE id = 32 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 34 | ประวัติศาสตร์เม็กซิโก ฉบับปรับปรุงใหม่ | ISBN 9786163016201
UPDATE books SET cover_url = 'https://api.chulabook.com/images/pid-15293.jpg'
WHERE id = 34 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 37 | ประวัติศาสตร์จีนฉบับย่อ | ISBN 9789740210580
UPDATE books SET cover_url = 'https://down-th.img.susercontent.com/file/th-11134207-7r98s-lz0f23v0iwmhff'
WHERE id = 37 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 40 | คู่มือเรียน เขียนโปรแกรม Python (ภาคปฏิบัติ) | ISBN 9786167502946
UPDATE books SET cover_url = 'https://down-th.img.susercontent.com/file/th-11134207-23030-0k90xvyqeoov04'
WHERE id = 40 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 41 | จัดการและวิเคราะห์ข้อมูลด้วย Python Data Science | ISBN 9786160839087
UPDATE books SET cover_url = 'https://down-th.img.susercontent.com/file/th-11134201-23030-fiaoxzb1nuov9a'
WHERE id = 41 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 42 | พื้นฐานการเขียนโปรแกรมด้วยภาษา Python | ISBN 9786164970953
UPDATE books SET cover_url = 'https://api.chulabook.com/images/pid-21160.JPG'
WHERE id = 42 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 43 | คู่มือเขียนโปรแกรมด้วยภาษาไพธอน Python | ISBN 9786162046971
UPDATE books SET cover_url = 'https://media.allonline.7eleven.co.th/pdmain/605079-00-programming-se-ed.jpg'
WHERE id = 43 AND (cover_url IS NULL OR btrim(cover_url) = '');

-- ID 48 | วิทยาศาสตร์ฉลาดรู้ เรื่องสัตว์พิษและพืชพิษ | ISBN 9789741080625
UPDATE books SET cover_url = 'https://down-th.img.susercontent.com/file/th-11134207-7r98o-lyj854qj2o3l17'
WHERE id = 48 AND (cover_url IS NULL OR btrim(cover_url) = '');

UPDATE books
SET cover_url = 'https://cdn.getbookie.com/product/image/2026/05/01/532x532/1777605065984_a173f3bb-7c17-53a0-9d55-c7a17ec8c08a.webp'
WHERE id = 22
  AND (cover_url IS NULL OR btrim(cover_url) = '');

UPDATE books
SET cover_url = 'https://cdn.getbookie.com/product/image/2023/09/07/full/1694061121115_3b8ae389-38b8-5370-a5f4-a5aab5075483.webp'
WHERE id = 24
  AND (cover_url IS NULL OR btrim(cover_url) = '');
COMMIT;

-- Check final cover count and show rows updated/available.
SELECT COUNT(*) AS books_with_covers
FROM books
WHERE cover_url IS NOT NULL AND btrim(cover_url) <> '';

SELECT id, title, isbn, cover_url
FROM books
WHERE id IN (3,6,7,9,10,11,13,16,18,19,21,22,24,32,34,37,40,41,42,43,48)
ORDER BY id;
