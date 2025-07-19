--
-- PostgreSQL database dump
--

-- Dumped from database version 15.12 (Debian 15.12-1.pgdg120+1)
-- Dumped by pg_dump version 15.12 (Debian 15.12-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: emelya
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO emelya;

--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: emelya
--

COMMENT ON SCHEMA public IS '';


--
-- Name: tarif_type; Type: TYPE; Schema: public; Owner: emelya
--

CREATE TYPE public.tarif_type AS ENUM (
    'Легкий старт',
    'Триумф',
    'Максимум'
);


ALTER TYPE public.tarif_type OWNER TO emelya;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: emelya
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO emelya;

--
-- Name: users; Type: TABLE; Schema: public; Owner: emelya
--

CREATE TABLE public.users (
    id integer NOT NULL,
    first_name text,
    last_name text,
    patronymic text,
    email text,
    phone text NOT NULL,
    is_email_verified boolean DEFAULT false,
    is_phone_verified boolean DEFAULT false,
    login text NOT NULL,
    password_hash text NOT NULL,
    referrer_id integer,
    card_number character varying(20),
    balance numeric(12,2) DEFAULT 0 NOT NULL,
    tarif public.tarif_type
);


ALTER TABLE public.users OWNER TO emelya;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: emelya
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO emelya;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: emelya
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: emelya
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: emelya
--

COPY public.schema_migrations (version, dirty) FROM stdin;
20250523	t
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: emelya
--

COPY public.users (id, first_name, last_name, patronymic, email, phone, is_email_verified, is_phone_verified, login, password_hash, referrer_id, card_number, balance, tarif) FROM stdin;
6	Атат	Оата	Вата	hj@gmail.com	+79779209532	f	t	imbJWsMB	$2a$10$xNef/hxOxqt4IDJofBt7M.FoeJPlahbSFCLDkKEFkbNLTlH2PEb6K	3	\N	0.00	\N
7	Dbfb	Brbe	Dbbd	gg@gmail.com	+79777791948	f	t	eISmfxZD	$2a$10$1/ZA6J.SNbG6TjVcopWH0eUkoVsBWECEked.tiHGvf7HfYUe7FLP2	3	\N	0.00	\N
8	Олеся	Миронова	Павловна	perevoda.master@yandex.ru	+79955012818	f	t	JMDuFSlr	$2a$10$g1sGRSyVgiUefHcEanDz5.xrziFBtJJwMD9j1ojX1F/UpPgLE6K9O	6	\N	0.00	\N
9	Hkcc	Hgg	Dyjkv	lady.cat.2014@yandex.ru	+79514060049	f	t	vsthLLkH	$2a$10$rw1Kf4MsyWj5HnM2x3Ul/O5qq72bvY2C4Z9LAp3.tLFSG7pZtTU6S	\N	\N	0.00	\N
12	Виталий	Чернов	Валерьевич	vital80@inbox.ru	+79255071744	f	t	RdZgaKXC	$2a$10$lEC7BSZpiMu/HVS7kjKcu.jEISRbhK5kLlI3hJPqnCAuBhcjB/p0i	\N	\N	0.00	\N
13	Елена	Логова	Игоревна	rus.sa.transit@gmail.com	+79912598661	f	t	VKJCukLm	$2a$10$Iw2RPLVpy9JHRh88om7Y4OOXkDMnpgqx3aHvfsFxjfPARppcZnAjq	\N	\N	0.00	\N
14	Кирилл	Лященко 	Андреевич	liashchenko94@gmail.com	+79657570424	f	t	UvafISup	$2a$10$9kOhm.eu.bYT1k7PtYhObep1Be21gQmH4VK5Dx1lvGgHJzkUL8tJK	\N	\N	0.00	\N
16	Тимур	Халилов	Русланович	jaxxxxx79@gmail.com	+79163037471	f	f	lMzjytJT	$2a$10$12AZKGxZSmW5krroKcLyheBaYiYe6Q5Nqo9cOPaaUhZpDitHmQCsG	\N	\N	0.00	\N
19	Иван	Иванов	Иванович	vovahdd9988@gmail.com	+79141863268	f	f	ZfivluUi	$2a$10$mIV2zNxfxEZlgfl4H6i2Gu/SykAbvKgYfHUP3JePtlc4KFZqguG32	\N	\N	0.00	\N
21	Иван	Иванов	Иванович	vovagd9988@gmail.com	+79241863268	f	f	DBsBZUew	$2a$10$23ojv3dvHAWkFFvcUMeK5uvxVEOygRfNU/7s.2h5mMlmoHjsadIVu	3	\N	0.00	\N
22	Дмитрий	Подолякин	Дмитриевич	playfix359@gmail.com	+79884837971	f	t	AChwRljB	$2a$10$.uLgpNYHTYRFCRC5QURfOOiHQuZxJ24v9m1FnRlw2I7024Dy29aAu	10	\N	0.00	\N
26	Иван	Котов	Кириллович	emel91962@gmail.com	+79779254087	f	t	RLRWgnwa	$2a$10$FtGjWPRbcecJ2BhjgDpML.Jq6rwkUmOM9..rNnNWTxCEiXV5rwRLu	\N	\N	0.00	\N
15	Виктор	Васильев	Викторович	victor309@mail.ru	+79267638822	f	t	pYSvvwNW	$2a$10$S0biFb13FnANV1D2UZgyh.J.JESwkwAVMxvDfF5F8AGob3aGOOCiS	\N	\N	5000.00	\N
23	Екатерина	Мейлихова	Викторовна	sk300884@bk.ru	+79771284622	f	t	bpZefAqA	$2a$10$DDqSr0cEMSNNYGcOK6SddOL0ogPKY3HxOeLnzv/akpmaLbVEqxGe.	\N	\N	2640000.00	\N
28	Дмитрий	Бреус	Германович 	Sportikbreus0099@gmail.com	+79003065601	f	t	bjTsLWcb	$2a$10$nV1LEichxbOYsYEYxydEF.AevyWuJxXfimLB2ppRInwaPSXHxmLaW	\N	\N	0.00	\N
24	Константин 	Прокудин 	Фёдорович	prokkonstanta@gmail.com	+79994570854	f	t	eAAqSALC	$2a$10$hsQaQzkTSgTz11E/jRXn5uJtExO16Jlii6QbE/Ww3MSZLfRJco8F6	\N	\N	62800.00	\N
27	Павел 	Романов	Максимович	perevoda.master@yandex.ru	+79856988950	f	t	SGxpyzsk	$2a$10$/TnBssfbwhifM99a8.dCLOVa7XrAZfCE8oZ.Mdv7mUBY4WDlIorsy	\N	\N	0.00	\N
3	Vova	Ushakov	sf	vovayhh9988@gmail.com	+79831863268	f	t	KJcUMnAA	$2a$10$UueprloAnWlGgFouNzQo6ez4ksk7PK/PlDMLj2gaf2WJs1CR1y0oK	\N	9876 5432 1345 6789	12.00	\N
10	Александр	Реужин	Артурович	alexreuzhin@gmail.com	+79164052188	f	t	Byluwbpk	$2a$10$XsNV7CT/b5DMSp2nxjpU7.odiSvv6b4k.eYmOK44GJ0gIS4bj.smi	22	\N	1191860.00	\N
25	Иван	Котов	Кириллович	emel91962@gmail.com	+79163955683	f	f	uLsxVKEM	$2a$10$o6xiBCVJ25/wQJpcFHby.uJPOAKpnZ88cYOOBiEZ8Jv.p05eEBXvm	\N	\N	0.00	\N
17	Василий	Бородин	Иванович	rus.usa.transit@gmail.com	+79309283104	f	t	wCmbIWay	$2a$10$kjS0Fl15RlHTh8V0GW4XMe2lHTfbjgGYJfwqxISV4MMWnh6wlXiHC	\N	\N	1463000.00	\N
\.


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: emelya
--

SELECT pg_catalog.setval('public.users_id_seq', 28, true);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: emelya
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: users users_login_key; Type: CONSTRAINT; Schema: public; Owner: emelya
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_login_key UNIQUE (login);


--
-- Name: users users_phone_key; Type: CONSTRAINT; Schema: public; Owner: emelya
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_phone_key UNIQUE (phone);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: emelya
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_referrer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: emelya
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_referrer_id_fkey FOREIGN KEY (referrer_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: emelya
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;


--
-- PostgreSQL database dump complete
--

