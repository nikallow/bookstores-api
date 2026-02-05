-- +goose Up
-- +goose StatementBegin
CREATE TABLE stores
(
    id         BIGSERIAL PRIMARY KEY,
    uuid       UUID        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    address    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL        DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL        DEFAULT now(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE books
(
    id               BIGSERIAL PRIMARY KEY,
    isbn             VARCHAR(13) NULL UNIQUE,
    title            TEXT        NOT NULL,
    author           TEXT        NOT NULL,
    description      TEXT        NULL,
    page_count       INTEGER     NULL CHECK (page_count > 0),
    publication_year INTEGER     NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ NULL
);

CREATE TABLE skus
(
    id              BIGSERIAL PRIMARY KEY,
    uuid            UUID        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    book_id         BIGINT      NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    store_id        BIGINT      NOT NULL REFERENCES stores (id) ON DELETE CASCADE,
    price_in_kopeks INTEGER     NOT NULL CHECK (price_in_kopeks >= 0),
    stock_count     INTEGER     NOT NULL        DEFAULT 0 CHECK (stock_count >= 0),
    created_at      TIMESTAMPTZ NOT NULL        DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL        DEFAULT now(),
    deleted_at      TIMESTAMPTZ NULL,
    UNIQUE (book_id, store_id)
);

INSERT INTO stores (name, address)
VALUES ('Магаз 1', 'г. Москва, ул. Тестовая, д. 1'),
       ('Магаз 1', 'г. Москва, ул. Тестовая, д. 2');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS skus;
DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS stores;
-- +goose StatementEnd