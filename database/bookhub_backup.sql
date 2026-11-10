--
-- PostgreSQL database dump
--

\restrict erwSEr7NVDROXRamE1939QaKfdbRf0iv8D3mS9EVqAebuGKv2IXnt4xdgmSE0aw

-- Dumped from database version 18.6
-- Dumped by pg_dump version 18.6

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: votes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.votes (
    member_id bigint NOT NULL,
    book_id bigint NOT NULL,
    score smallint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT votes_score_check CHECK (((score >= 1) AND (score <= 10)))
);


--
-- Name: book_vote_stats; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.book_vote_stats AS
 SELECT book_id,
    count(*) AS vote_count,
    round(avg(score), 1) AS vote_average
   FROM public.votes
  GROUP BY book_id;


--
-- Name: books; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.books (
    id bigint NOT NULL,
    title character varying(500) NOT NULL,
    author character varying(255) NOT NULL,
    category character varying(100),
    isbn character varying(20),
    publisher character varying(255),
    publication_year integer,
    description text,
    price_thb numeric(10,2),
    stock_quantity integer DEFAULT 0 NOT NULL,
    cover_url character varying(500),
    source_url character varying(500),
    language character(2) DEFAULT 'th'::bpchar NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT books_stock_quantity_check CHECK ((stock_quantity >= 0))
);


--
-- Name: books_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.books_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: books_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.books_id_seq OWNED BY public.books.id;


--
-- Name: members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.members (
    id bigint NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    display_name character varying(50) NOT NULL,
    avatar_url character varying(500),
    role character varying(10) DEFAULT 'member'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT members_display_name_check CHECK ((char_length(btrim((display_name)::text)) >= 1)),
    CONSTRAINT members_role_check CHECK (((role)::text = ANY ((ARRAY['member'::character varying, 'admin'::character varying])::text[])))
);


--
-- Name: members_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: members_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.members_id_seq OWNED BY public.members.id;


--
-- Name: reviews; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reviews (
    id bigint NOT NULL,
    member_id bigint NOT NULL,
    book_id bigint NOT NULL,
    text text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reviews_text_check CHECK ((char_length(btrim(text)) >= 10))
);


--
-- Name: reviews_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.reviews_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: reviews_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.reviews_id_seq OWNED BY public.reviews.id;


--
-- Name: wishlist_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.wishlist_items (
    member_id bigint NOT NULL,
    book_id bigint NOT NULL,
    note character varying(200),
    added_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: books id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.books ALTER COLUMN id SET DEFAULT nextval('public.books_id_seq'::regclass);


--
-- Name: members id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.members ALTER COLUMN id SET DEFAULT nextval('public.members_id_seq'::regclass);


--
-- Name: reviews id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews ALTER COLUMN id SET DEFAULT nextval('public.reviews_id_seq'::regclass);


--
-- Data for Name: books; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.books (id, title, author, category, isbn, publisher, publication_year, description, price_thb, stock_quantity, cover_url, source_url, language, created_at, updated_at) FROM stdin;
1	โฟร์ทวิง (Fourth Wing)	รีเบกกา ยาร์รอส	นวนิยายแฟนตาซี	9786161873554	\N	2025	ไวโอเล็ตต้องเข้าเรียนในวิทยาลัยฝึกมังกรที่เต็มไปด้วยการแข่งขันและอันตราย เพื่อพิสูจน์ว่าตนเองคู่ควรกับการเป็นนักขี่มังกร	845.00	12	https://books.google.com/books/content?id=EP-gEAAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.chulabook.com/literature/216455	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:42:01.038235+07
2	ใช้คลื่นพลังบวกดึงดูดพลังสุข (Good Vibes, Good Life)	เว็กซ์ คิงส์	จิตวิทยาและพัฒนาตนเอง	9786161844899	อมรินทร์ฮาวทู	2023	แนวทางดูแลตัวเอง ปรับมุมมอง และสร้างนิสัยเชิงบวกเพื่อให้ใช้ชีวิตอย่างมีความสุขและมั่นใจมากขึ้น	245.00	18	https://books.google.com/books/content?id=3wBlDwAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.chulabook.com/psychology-development/135854	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:42:08.722221+07
4	ทฤษฎีปล่อยเขา (The Let Them Theory)	เมล รอบบินส์	จิตวิทยาและพัฒนาตนเอง	9786161877026	ฮาวทู	2025	เสนอแนวคิดให้หยุดพยายามควบคุมการกระทำของคนอื่น แล้วหันมาโฟกัสสิ่งที่ตนเองเลือกและจัดการได้	395.00	16	https://books.google.com/books/content?id=PZM4EQAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.nstda.or.th/library/opac/Book/55048?c=1675751124	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:42:21.868236+07
5	เรื่องที่ไม่มีใครสอนเราเกี่ยวกับความรัก (Things No One Taught Us About Love)	เว็กซ์ คิงส์	จิตวิทยาและความสัมพันธ์	9786161873790	ฮาวทู	2025	ชวนสำรวจความรัก ความสัมพันธ์ และการเห็นคุณค่าในตัวเอง ผ่านมุมมองด้านจิตวิทยาและประสบการณ์ชีวิต	295.00	10	https://books.google.com/books/content?id=OOPg0AEACAAJ&printsec=frontcover&img=1&zoom=1&source=gbs_api	https://www.botlc.or.th/item/book/01000081105	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:42:33.485297+07
8	ฉันตกลงสงบศึกกับจิตใจ	สึเนโกะ นากามูระ (Tsuneko Nakamura)	จิตวิทยาและพัฒนาตนเอง	9786160468201	นานมีบุ๊คส์	2024	ชวนทำความเข้าใจความคิดและความรู้สึกของตัวเอง พร้อมฝึกอยู่กับใจอย่างอ่อนโยนและลดการต่อสู้กับอารมณ์	265.00	15	https://books.google.com/books/content?id=97syEQAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.naiin.com/product/detail/621675	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:42:38.896097+07
12	เราต่างหมดไฟ ในสังคมคลั่งไคล้การทำงาน (The Burnout Society)	บย็อง-ช็อล ฮัน	จิตวิทยาและสังคม	9786169446453	สำนักพิมพ์สามย่าน	2024	วิเคราะห์สังคมที่ผลักให้ผู้คนต้องทำงานและพัฒนาตัวเองตลอดเวลา จนเกิดความเหนื่อยล้าและหมดไฟ	250.00	13	https://books.google.com/books/content?id=SQoNCgAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://samyanpress.bentoweb.com/th/product/1136570/product-1136570	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:42:44.405529+07
14	จิตวิทยาเอาชนะอนาคต (Same as Ever)	มอร์แกน เฮาเซิล	จิตวิทยาและพัฒนาตนเอง	9786168187661	ลีฟ ริช ฟอร์เอฟเวอร์	2024	ชวนมองสิ่งที่ไม่ค่อยเปลี่ยนแปลงในพฤติกรรมมนุษย์และสังคม เพื่อใช้เป็นหลักคิดในการตัดสินใจและวางแผนอนาคต	350.00	12	https://books.google.com/books/content?id=-xrcEAAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.botlc.or.th/item/book/01000081048	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:42:50.245349+07
15	จิตวิทยา: ความรู้ฉบับพกพา (Psychology: A Very Short Introduction)	จิลเลียน บัตเลอร์ และ เฟรดา แม็กมานัส	จิตวิทยาและพัฒนาตนเอง	9786167885209	โอเพ่นเวิลด์ส	2015	แนะนำแนวคิดพื้นฐานทางจิตวิทยาและวิธีที่นักจิตวิทยาใช้ศึกษาอารมณ์ ความคิด และพฤติกรรมของมนุษย์	220.00	5	https://books.google.com/books/content?id=dGJoAgAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.botlc.or.th/item/book/01000077117	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:42:55.653158+07
20	อย่า อยู่ อย่าง ยาก (How to Do Hard Things)	เวโรนิกา เดียร์ลี	จิตวิทยาและพัฒนาตนเอง	9786160463589	บลูม พับลิชชิ่ง	2024	แนวทางรับมือกับเรื่องยาก ฝึกความกล้า และค่อย ๆ ลงมือทำสิ่งที่ท้าทายแทนการหลีกเลี่ยง	185.00	20	https://books.google.com/books/content?id=xLGOzgEACAAJ&printsec=frontcover&img=1&zoom=1&source=gbs_api	https://www.botlc.or.th/item/book/01000081033	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:43:07.623119+07
23	ไม่ต้องได้ทุกอย่างเพื่อมีความสุข (JOMO)	ธนา เธียรอัจฉริยะ	จิตวิทยาและการใช้ชีวิต	9786168350027	Koob	2024	ชวนลดความกดดันจากการต้องตามทุกกระแส และค้นหาความสุขจากการเลือกสิ่งสำคัญกับชีวิตตัวเอง	320.00	18	https://books.google.com/books/content?id=nLPtQ6Fpv9IC&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.botlc.or.th/item/book/01000081043	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:43:11.72771+07
25	ความมั่งคั่งที่เงินซื้อไม่ได้ (The Wealth Money Can't Buy)	โรบิน ชาร์มา	จิตวิทยาและการใช้ชีวิต	9786162877858	วีเลิร์น	2025	นำเสนอความมั่งคั่งในหลายมิติ ทั้งสุขภาพ ความสัมพันธ์ การเติบโตภายใน และความหมายของชีวิต	390.00	8	https://books.google.com/books/content?id=JyfgEAAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.botlc.or.th/item/book/01000081314	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:43:27.184109+07
26	Manifest: Dive Deeper ดำดิ่งลึกซึ้ง ไปให้ถึงทุกปรารถนา	ร็อกซี นาฟูรี	จิตวิทยาและพัฒนาตนเอง	9786161865450	Howto	2024	ต่อยอดแนวคิดการตั้งเป้าหมายและการลงมือทำ เพื่อให้ผู้อ่านทบทวนความต้องการและสร้างเส้นทางสู่เป้าหมาย	275.00	7	https://books.google.com/books/content?id=WSV0EAAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.botlc.or.th/item/book/01000081011	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:43:32.105995+07
27	มองเห็นพลังล่องหนที่ควบคุมจิตใจคน (Invisible Influence)	โจนาห์ เบอร์เกอร์	จิตวิทยาและการตลาด	9786162877308	วีเลิร์น	2024	อธิบายอิทธิพลทางสังคมที่ส่งผลต่อการตัดสินใจ รสนิยม และพฤติกรรมของผู้คนโดยที่เราอาจไม่รู้ตัว	280.00	6	https://books.google.com/books/content?id=rNrPCgAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.botlc.or.th/item/book/01000081106	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:43:36.389611+07
28	เท่าทันอารมณ์ก็เข้าใจตนเอง (Emotional Agility)	ซูซาน เดวิด	จิตวิทยาและพัฒนาตนเอง	9786168293881	บิบลิโอ	2023	แนะนำวิธีรับรู้อารมณ์และความคิดของตัวเองโดยไม่ปล่อยให้สิ่งเหล่านั้นกำหนดการตัดสินใจทั้งหมด	340.00	9	https://books.google.com/books/content?id=5HLNDAAAQBAJ&printsec=frontcover&img=1&zoom=1&source=gbs_api	https://www.botlc.or.th/item/book/01000081139	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:43:40.688246+07
29	ชีวิตที่ดีมาก ๆ: คุณค่าของความล้มเหลว และความสำคัญของจินตนาการ (Very Good Lives)	เจ.เค. โรว์ลิ่ง	แรงบันดาลใจ	9786160437672	นานมีบุ๊คส์	2017	ถ่ายทอดแนวคิดเรื่องความล้มเหลว จินตนาการ และการค้นหาศักยภาพของตัวเองจากคำบรรยายสร้างแรงบันดาลใจ	180.00	6	https://books.google.com/books/content?id=QIgbEAAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.botlc.or.th/item/book/01000080156	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:43:45.750576+07
30	การจัดการการเงินธุรกิจ (Fundamentals of Corporate Finance)	Stephen A. Ross, Randolph W. Westerfield และ Bradford D. Jordan	ธุรกิจและการเงิน	9789749965214	แมคกรอ-ฮิล	2014	ตำราพื้นฐานด้านการเงินองค์กร ครอบคลุมการตัดสินใจลงทุน การจัดหาเงินทุน และการบริหารการเงินของธุรกิจ	590.00	4	https://books.google.com/books/content?id=knxzLw1se7sC&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.botlc.or.th/item/book/01000073942	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:44:03.673161+07
31	จัดการ Cash ก่อน Crash	ธนเดช มหโภไคย	ธุรกิจและการเงิน	9786167126913	ธรรมนิติ เพรส	2012	อธิบายแนวคิดการจัดการกระแสเงินสดและการเงินธุรกิจเพื่อช่วยลดความเสี่ยงด้านสภาพคล่อง	390.00	5	https://books.google.com/books/content?id=X11VDwAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.botlc.or.th/item/book/01000068504	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:44:04.323276+07
33	ทำธุรกิจ เริ่มจาก 10 ให้ได้ 100 (The 10% Entrepreneur)	Patrick J. McGinnis	ธุรกิจและผู้ประกอบการ	9786161828905	Shortcut	2019	นำเสนอแนวคิดการเริ่มทำธุรกิจควบคู่กับงานประจำ โดยใช้เวลาและทรัพยากรอย่างเหมาะสม	260.00	8	https://books.google.com/books/content?id=xfStjgEACAAJ&printsec=frontcover&img=1&zoom=1&source=gbs_api	https://www.botlc.or.th/item/book/01000079305	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:44:11.11725+07
35	ประวัติศาสตร์ไทยสมัยธนบุรี	ก่องแก้ว วีระประจักษ์ และกรมศิลปากร	ประวัติศาสตร์ไทย	9786162834110	กรมศิลปากร	2018	นำเสนอเหตุการณ์ บุคคล และพัฒนาการสำคัญของสยามในช่วงกรุงธนบุรี	450.00	4	https://books.google.com/books/content?id=KoZ2i2Lr7pkC&printsec=frontcover&img=1&zoom=1&source=gbs_api	https://www.botlc.or.th/item/book/01000078999	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:44:15.235529+07
36	ประวัติศาสตร์ไทยฉบับสังเขป (Thailand: A Short History)	David K. Wyatt	ประวัติศาสตร์ไทย	9786167202532	มูลนิธิโครงการตำราสังคมศาสตร์และมนุษยศาสตร์	2014	ภาพรวมประวัติศาสตร์ไทยตั้งแต่ยุคแรกเริ่มถึงสังคมสมัยใหม่ พร้อมบริบททางการเมืองและวัฒนธรรม	390.00	5	https://books.google.com/books/content?id=8YBujWXwkakC&printsec=frontcover&img=1&zoom=1&source=gbs_api	https://www.botlc.or.th/item/book/01000073820	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:44:19.182754+07
38	ประวัติศาสตร์ไทยร่วมสมัย	คริส เบเคอร์ และ ผาสุก พงษ์ไพจิตร	ประวัติศาสตร์ไทย	9789740212652	มติชน	2014	วิเคราะห์การเปลี่ยนแปลงทางการเมือง เศรษฐกิจ และสังคมไทยในยุคร่วมสมัย	390.00	7	https://books.google.com/books/content?id=rYpZDwAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.botlc.or.th/item/book/01000073789	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:44:19.836551+07
39	ประวัติศาสตร์ชาติไทย	ธารณา คชเสนี และคณะ	ประวัติศาสตร์ไทย	9786169245001	เพ็ชรสายสวัสดิ์	2016	รวบรวมเรื่องราวและเหตุการณ์สำคัญที่เกี่ยวข้องกับพัฒนาการของชาติไทย พร้อมภาพประกอบและแผนที่	450.00	4	https://books.google.com/books/content?id=DdYf23hADJsC&printsec=frontcover&img=1&zoom=1&source=gbs_api	https://www.botlc.or.th/item/book/01000076425	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:44:23.36262+07
44	ประวัติศาสตร์ว่าด้วยเรื่องเงิน (Money)	เจคอบ โกลด์สไตน์	ประวัติศาสตร์เศรษฐกิจและการเงิน	9786164343788	แอร์โรว์	2024	เล่าพัฒนาการของเงินและระบบการเงิน ผ่านเหตุการณ์และแนวคิดที่เปลี่ยนแปลงเศรษฐกิจโลก	290.00	9	https://books.google.com/books/content?id=rEOZQgAACAAJ&printsec=frontcover&img=1&zoom=1&source=gbs_api	https://www.botlc.or.th/item/book/01000081050	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:44:29.494802+07
46	วิทยาศาสตร์ฉลาดรู้ เรื่อง ที่สุดของโลก	ชาฮยอนจิน	วิทยาศาสตร์และเยาวชน	9786165180146	อมรินทร์คอมมิกส์	2009	นำเสนอเกร็ดความรู้ทางวิทยาศาสตร์เกี่ยวกับสิ่งที่มีความโดดเด่นและเป็นที่สุดในโลกในรูปแบบอ่านง่าย	180.00	7	https://cdn.getbookie.com/product/image/2025/01/13/532x532/1736776059047_aa80c754-349d-5ec9-bdde-a28cc5d65c46.webp	https://www.botlc.or.th/item/book/01000078693	th	2026-11-10 22:19:12.247706+07	2026-11-10 23:16:01.351272+07
3	Money 101: เริ่มต้นนับหนึ่งสู่ชีวิตการเงินอุดมสุข	จักรพงษ์ เมษพันธุ์ (The Money Coach)	การเงินและการลงทุน	9786160834563	ซีเอ็ดยูเคชั่น	2019	ปูพื้นฐานการจัดการเงินส่วนบุคคล ตั้งแต่การวางแผนรายรับรายจ่าย การออม และการจัดการเป้าหมายทางการเงิน	250.00	14	https://storage.naiin.com/system/application/bookstore/resource/product/201910/491276/1000225492_front_XXL.jpg?v=1725795565	https://www.botlc.or.th/item/book/01000079432	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
6	เดนดาว NEVER DIE	จิตจงกล	นวนิยายไทย	9786168166956	พี.เอส.	2025	นวนิยายที่เล่าเรื่องชีวิต ความทรงจำ ความเหงา และความรัก ผ่านตัวละครที่พยายามประคองชีวิตท่ามกลางอดีตที่ซับซ้อน	380.00	8	https://api.chulabook.com/images/pid-231417.JPG	https://dev.chulabook.com/literature/231417	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
7	52 เฮิรตซ์...คลื่นเสียงที่ไม่มีใครได้ยิน	มาจิดะ โซโนะโกะ	นวนิยายแปล	9786161846985	Piccolo	2022	เรื่องราวของหญิงสาวและเด็กชายที่ต่างเคยถูกทำร้ายและโดดเดี่ยว ก่อนจะได้พบกันและค่อย ๆ เรียนรู้ที่จะรับฟังกัน	295.00	11	https://storage.naiin.com/system/application/bookstore/resource/product/202202/541412/1000246409_front_XXL.jpg?v=1725894189	https://www.naiin.com/product/detail/541412	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
9	ไนปุรานเริกส์	กิตติศักดิ์ คงคา	นวนิยายไทย	9786168296578	13357 Publishing	2026	นวนิยายไทยร่วมสมัยที่พาผู้อ่านติดตามเรื่องราวและความสัมพันธ์ของตัวละครผ่านเหตุการณ์ที่คาดเดาได้ยาก	300.00	9	https://storage.naiin.com/system/application/bookstore/resource/product/202604/701156/1000290052_front_XXL.jpg?v=1775125814	https://www.naiin.com/product/detail/701156	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
11	วิถีแห่งการลงทุนของผม บทเรียนจากเซียนหุ้นญี่ปุ่น	ทาซึโร่ คิโยฮาระ	การเงินและการลงทุน	9786169440031	MPIRIKABOOKS	2025	ถ่ายทอดประสบการณ์และแนวคิดการลงทุนในตลาดหุ้นญี่ปุ่น พร้อมบทเรียนเกี่ยวกับการประเมินธุรกิจและการบริหารพอร์ต	439.00	6	https://down-th.img.susercontent.com/file/th-11134207-7rash-m4stjsr2gud2e1	https://www.phanpha.com/item/%E0%B8%A7%E0%B8%B4%E0%B8%96%E0%B8%B5%E0%B9%81%E0%B8%AB%E0%B9%88%E0%B8%87%E0%B8%81%E0%B8%B2%E0%B8%A3%E0%B8%A5%E0%B8%87%E0%B8%97%E0%B8%B8%E0%B8%99%E0%B8%82%E0%B8%AD%E0%B8%87%E0%B8%9C%E0%B8%A1-%E0%B8%9A%E0%B8%97%E0%B9%80%E0%B8%A3%E0%B8%B5%E0%B8%A2%E0%B8%99%E0%B8%88%E0%B8%B2%E0%B8%81%E0%B9%80%E0%B8%8B%E0%B8%B5%E0%B8%A2%E0%B8%99%E0%B8%AB%E0%B8%B8%E0%B9%89%E0%B8%99%E0%B8%8D%E0%B8%B5%E0%B9%88%E0%B8%9B%E0%B8%B8%E0%B9%88%E0%B8%99	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
13	คำสอนจากมหาเศรษฐีแห่งบาบิโลน	จอร์จ เอส. คลาสัน	การเงินและการลงทุน	9786168335819	วารา	2025	เล่าแนวคิดการออม การรักษาเงิน และการสร้างความมั่งคั่งผ่านเรื่องเล่าจากนครบาบิโลนโบราณ	395.00	17	https://storage.naiin.com/system/application/bookstore/resource/product/202501/634674/1000278786_front_XXL.jpg?v=1736410577	https://www.naiin.com/product/detail/634674	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
50	เปนชู้กับผี	วิศิษฏ์ ศาสนเที่ยง	นวนิยายสยองขวัญ	9786160051939	SHADOW	2024	นวนิยายสยองขวัญเกี่ยวกับหญิงสาวที่เดินทางเข้าพระนครเพื่อตามหาคนรัก ก่อนพบความลับประหลาดในบ้านที่เธอเข้าไปพัก	210.00	6	https://books.google.com/books/content?id=sYwzEQAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api	https://www.naiin.com/product/detail/629286	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:45:08.512991+07
10	Longevity by Design ศาสตร์แห่งการออกแบบสุขภาพที่ทำให้อายุยืนยาว	นรินทร สุรสินธน	สุขภาพและการดูแลตัวเอง	9786169442172	ไรเตอร์โซล	2024	อธิบายแนวคิดการดูแลสุขภาพระยะยาว ทั้งพันธุกรรม เวชศาสตร์วิถีชีวิต สิ่งแวดล้อม และการวางแผนคุณภาพชีวิต	380.00	7	https://storage.naiin.com/system/application/bookstore/resource/product/202411/631777/1000277948_front_XXL.jpg?v=1732802464	https://www.naiin.com/product/detail/631777	th	2026-11-10 22:19:12.247706+07	2026-11-10 23:16:01.351272+07
16	จัดการคนรับมือยาก 9 ไทป์ยังไงให้อยู่หมัด	นิก โรบินสัน	จิตวิทยาและการสื่อสาร	9786160474141	นามิ พับลิชชิ่ง	2026	แนะนำวิธีเข้าใจรูปแบบพฤติกรรมของคนที่รับมือยากและเลือกวิธีสื่อสารให้เหมาะกับสถานการณ์	320.00	8	https://platform-api.nanmeebooks.com/uploads/images/image-1769160429894.png	https://www.botlc.or.th/item/book/01000081361	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
18	Confidence: 8 ขั้นตอนสร้างความมั่นใจที่แท้จริง	ร็อกซี นาฟูรี	จิตวิทยาและพัฒนาตนเอง	9786161885434	Howto	2025	เสนอแบบฝึกและแนวทางสร้างความเชื่อมั่นในตัวเองผ่านการปรับความคิดและลงมือทำอย่างต่อเนื่อง	295.00	16	https://storage.naiin.com/system/application/bookstore/resource/product/202512/691815/1000287475_front_XXL.jpg?v=1765420328	https://www.botlc.or.th/item/book/01000081352	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
19	จิตวิทยาแค่ 1% ทำให้คุณเหนือคน	เฌอมาณย์ รัตนพงศ์ตระกูล	จิตวิทยาและพัฒนาตนเอง	9786163651365	ทัช	2015	รวบรวมแนวคิดด้านจิตวิทยาและการพัฒนาตัวเองเพื่อช่วยให้คิดเป็นระบบและสร้างความได้เปรียบในชีวิตประจำวัน	199.00	9	https://down-th.img.susercontent.com/file/36084ebc3638ea10d9c9e5404d73da82	https://www.botlc.or.th/item/book/01000077044	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
32	การเงิน 101	พี่ทุย (Money Buffalo)	การเงินและการลงทุน	9786160837564	ซีเอ็ดยูเคชั่น	2020	แนะนำพื้นฐานการเงินส่วนบุคคลและการลงทุนด้วยภาษาที่เข้าใจง่าย เหมาะสำหรับผู้เริ่มต้น	220.00	17	https://down-th.img.susercontent.com/file/th-11134207-7qul3-lf0wlh9nep0157	https://www.botlc.or.th/item/book/01000079607	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
34	ประวัติศาสตร์เม็กซิโก ฉบับปรับปรุงใหม่	ปาโบล เอสกาลันเต กอนซาลโบ	ประวัติศาสตร์	9786163016201	ยิปซี กรุ๊ป	2017	สรุปพัฒนาการทางประวัติศาสตร์ของเม็กซิโก ตั้งแต่ยุคก่อนสมัยใหม่จนถึงประเด็นสำคัญในประวัติศาสตร์ร่วมสมัย	420.00	4	https://api.chulabook.com/images/pid-15293.jpg	https://www.botlc.or.th/item/book/01000080056	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
37	ประวัติศาสตร์จีนฉบับย่อ	หลี่เฉวียน	ประวัติศาสตร์โลก	9789740210580	มติชน	2014	เรียบเรียงเหตุการณ์และการเปลี่ยนแปลงสำคัญของจีนในแต่ละยุคสมัย	420.00	6	https://down-th.img.susercontent.com/file/th-11134207-7r98s-lz0f23v0iwmhff	https://www.botlc.or.th/item/book/01000073814	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
40	คู่มือเรียน เขียนโปรแกรม Python (ภาคปฏิบัติ)	โชติพันธุ์ หล่อเลิศสุนทร และ ฐิตะพันธุ์ หล่อเลิศสุนทร	เทคโนโลยีและการเขียนโปรแกรม	9786167502946	คอร์ฟังก์ชั่น	2019	คู่มือฝึกเขียนโปรแกรมภาษา Python ผ่านตัวอย่างและแบบฝึกหัดเชิงปฏิบัติ	350.00	6	https://down-th.img.susercontent.com/file/th-11134207-23030-0k90xvyqeoov04	https://www.botlc.or.th/item/book/01000079501	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
41	จัดการและวิเคราะห์ข้อมูลด้วย Python Data Science	บัญชา ปะสีละเตสัง	เทคโนโลยีและการเขียนโปรแกรม	9786160839087	ซีเอ็ดยูเคชั่น	2020	แนะนำการใช้ Python เพื่อจัดการ วิเคราะห์ และนำเสนอข้อมูลด้วยแนวทางด้าน Data Science	390.00	7	https://down-th.img.susercontent.com/file/th-11134201-23030-fiaoxzb1nuov9a	https://www.botlc.or.th/item/book/01000080750	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
42	พื้นฐานการเขียนโปรแกรมด้วยภาษา Python	ประยงค์ อู่ประสิทธิ์วงศ์	เทคโนโลยีและการเขียนโปรแกรม	9786164970953	ส. เอเซียเพรส (1989)	2019	อธิบายพื้นฐานภาษา Python ตั้งแต่ตัวแปร เงื่อนไข การวนซ้ำ ไปจนถึงการเขียนโปรแกรมเบื้องต้น	320.00	5	https://api.chulabook.com/images/pid-21160.JPG	https://www.botlc.or.th/item/book/01000079498	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
48	วิทยาศาสตร์ฉลาดรู้ เรื่อง สัตว์พิษและพืชพิษ	ชินเฮยอง	วิทยาศาสตร์และเยาวชน	9789741080625	อมรินทร์คอมมิกส์	2008	ให้ความรู้เกี่ยวกับสัตว์และพืชที่มีพิษ พร้อมข้อมูลพื้นฐานเพื่อเรียนรู้และระมัดระวังอันตรายจากธรรมชาติ	180.00	5	https://down-th.img.susercontent.com/file/th-11134207-7r98o-lyj854qj2o3l17	https://www.botlc.or.th/item/book/01000078001	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
22	ไม่ยากถ้าอยากเก่งอินเทอร์เน็ต	Park Jee-Won และ Kim Do-Young	หนังสือเด็กและเยาวชน	9786160404612	นานมีบุ๊คส์	2011	หนังสือความรู้สำหรับเยาวชนที่แนะนำการใช้อินเทอร์เน็ตและแนวคิดพื้นฐานด้านเทคโนโลยี	165.00	5	https://cdn.getbookie.com/product/image/2026/05/01/532x532/1777605065984_a173f3bb-7c17-53a0-9d55-c7a17ec8c08a.webp	https://www.botlc.or.th/item/book/01000079931	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
24	เริ่มจาก 1	อันยา มงคล และ วิสูตร แสงอรุณเลิศ	จิตวิทยาและพัฒนาตนเอง	9786167752549	สต็อคทูมอร์โรว์	2015	กระตุ้นให้เริ่มต้นเปลี่ยนแปลงจากสิ่งเล็ก ๆ และลงมือทำอย่างเป็นขั้นตอนเพื่อไปสู่เป้าหมาย	220.00	10	https://cdn.getbookie.com/product/image/2023/09/07/full/1694061121115_3b8ae389-38b8-5370-a5f4-a5aab5075483.webp	https://www.botlc.or.th/item/book/01000074135	th	2026-11-10 22:19:12.247706+07	2026-11-10 22:19:12.247706+07
21	คุณในอนาคต ต้องขอบคุณตัวเอง ที่อดทนสู้ในวันนี้	ทัง มู่	จิตวิทยาและพัฒนาตนเอง	9789749760970	บี มีเดีย	2017	ชวนสร้างวินัยและความอดทนในปัจจุบัน เพื่อให้ตัวเองในอนาคตได้รับผลจากการตัดสินใจที่ดี	220.00	11	https://storage.naiin.com/system/application/bookstore/resource/product/201705/214785/1000197225_front_XXL.jpg?v=1725848179	https://www.botlc.or.th/item/book/01000076561	th	2026-11-10 22:19:12.247706+07	2026-11-10 23:16:01.351272+07
43	คู่มือเขียนโปรแกรมด้วยภาษาไพธอน Python	ณัฐวัตร คำภักดี	เทคโนโลยีและการเขียนโปรแกรม	9786162046971	Provision	2018	คู่มือเรียนรู้การเขียนโปรแกรม Python พร้อมตัวอย่างการใช้งานและแนวทางฝึกเขียนโค้ด	330.00	5	https://cdn-ookbee.okbcdn.net/Books/Provision/2019/20191108072735/Thumbnails/Cover.jpg	https://www.botlc.or.th/item/book/01000078289	th	2026-11-10 22:19:12.247706+07	2026-11-10 23:16:01.351272+07
\.


--
-- Data for Name: members; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.members (id, email, password_hash, display_name, avatar_url, role, created_at, updated_at) FROM stdin;
1	ww@ww.ww	$2a$10$iCylLI5WKuu6syj.qgbszu2X.Nc/YBWSYBOZveJabch5nBnOc29F.	qweqwewqe	\N	member	2026-11-10 22:23:34.264256+07	2026-11-10 22:23:34.264256+07
2	admin@a.a	$2a$10$BBXvZ8Yz6ojByg0awBOwcOL9LS9oadMMGHccXvomPq115OgNoqWvy	admin	\N	admin	2026-11-10 23:21:12.92515+07	2026-11-10 23:24:21.453834+07
\.


--
-- Data for Name: reviews; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.reviews (id, member_id, book_id, text, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: votes; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.votes (member_id, book_id, score, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: wishlist_items; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.wishlist_items (member_id, book_id, note, added_at) FROM stdin;
\.


--
-- Name: books_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.books_id_seq', 50, true);


--
-- Name: members_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.members_id_seq', 2, true);


--
-- Name: reviews_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.reviews_id_seq', 1, false);


--
-- Name: books books_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.books
    ADD CONSTRAINT books_pkey PRIMARY KEY (id);


--
-- Name: members members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.members
    ADD CONSTRAINT members_pkey PRIMARY KEY (id);


--
-- Name: reviews reviews_member_id_book_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_member_id_book_id_key UNIQUE (member_id, book_id);


--
-- Name: reviews reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_pkey PRIMARY KEY (id);


--
-- Name: votes votes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.votes
    ADD CONSTRAINT votes_pkey PRIMARY KEY (member_id, book_id);


--
-- Name: wishlist_items wishlist_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wishlist_items
    ADD CONSTRAINT wishlist_items_pkey PRIMARY KEY (member_id, book_id);


--
-- Name: idx_books_author_lower; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_books_author_lower ON public.books USING btree (lower((author)::text));


--
-- Name: idx_books_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_books_category ON public.books USING btree (category);


--
-- Name: idx_books_title_lower; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_books_title_lower ON public.books USING btree (lower((title)::text));


--
-- Name: idx_reviews_book_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_reviews_book_created ON public.reviews USING btree (book_id, created_at DESC);


--
-- Name: idx_votes_book; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_votes_book ON public.votes USING btree (book_id);


--
-- Name: idx_wishlist_member_added; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_wishlist_member_added ON public.wishlist_items USING btree (member_id, added_at DESC);


--
-- Name: uq_members_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_members_email ON public.members USING btree (lower((email)::text));


--
-- Name: reviews reviews_book_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_book_id_fkey FOREIGN KEY (book_id) REFERENCES public.books(id) ON DELETE CASCADE;


--
-- Name: reviews reviews_member_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_member_id_fkey FOREIGN KEY (member_id) REFERENCES public.members(id) ON DELETE CASCADE;


--
-- Name: votes votes_book_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.votes
    ADD CONSTRAINT votes_book_id_fkey FOREIGN KEY (book_id) REFERENCES public.books(id) ON DELETE CASCADE;


--
-- Name: votes votes_member_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.votes
    ADD CONSTRAINT votes_member_id_fkey FOREIGN KEY (member_id) REFERENCES public.members(id) ON DELETE CASCADE;


--
-- Name: wishlist_items wishlist_items_book_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wishlist_items
    ADD CONSTRAINT wishlist_items_book_id_fkey FOREIGN KEY (book_id) REFERENCES public.books(id) ON DELETE CASCADE;


--
-- Name: wishlist_items wishlist_items_member_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.wishlist_items
    ADD CONSTRAINT wishlist_items_member_id_fkey FOREIGN KEY (member_id) REFERENCES public.members(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict erwSEr7NVDROXRamE1939QaKfdbRf0iv8D3mS9EVqAebuGKv2IXnt4xdgmSE0aw

