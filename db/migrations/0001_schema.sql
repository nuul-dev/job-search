-- +goose Up

CREATE TABLE users (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email      text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE profiles (
    user_id             bigint PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    priorities          jsonb,
    search_queries      text[],
    stack               text[],
    location            text,
    work_format         jsonb,
    blacklist_companies text[],
    exclusions          jsonb,
    updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE vacancies (
    id                 bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    source             text NOT NULL,
    source_external_id text NOT NULL,
    url                text NOT NULL,
    title              text NOT NULL,
    company            text,
    salary             text,
    location           text,
    work_format        text,
    description        text,
    raw                jsonb,
    fetched_at         timestamptz NOT NULL DEFAULT now(),
    published_at       timestamptz,
    UNIQUE (source, source_external_id)
);

CREATE TABLE vacancy_scores (
    vacancy_id bigint NOT NULL REFERENCES vacancies(id) ON DELETE CASCADE,
    user_id    bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score      integer NOT NULL,
    group_name text,
    notes      text,
    ranked_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (vacancy_id, user_id)
);

CREATE TABLE applications (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    vacancy_id bigint REFERENCES vacancies(id) ON DELETE SET NULL,
    user_id    bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status     text NOT NULL DEFAULT 'draft' CONSTRAINT applications_status_check CHECK (status IN ('draft','sent','replied','screen','test','offer','rejected','archived')),
    draft_body text,
    sent_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE application_messages (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    application_id bigint NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    author         text NOT NULL CONSTRAINT application_messages_author_check CHECK (author IN ('me','recruiter')),
    body           text NOT NULL,
    sent_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE fetch_runs (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    started_at    timestamptz NOT NULL DEFAULT now(),
    finished_at   timestamptz,
    source        text NOT NULL,
    fetched_count integer,
    error         text
);

CREATE INDEX ON vacancies (fetched_at DESC);
CREATE INDEX ON vacancy_scores (user_id, score DESC);
CREATE INDEX ON applications (user_id, status);

-- +goose Down

DROP TABLE IF EXISTS application_messages;
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS vacancy_scores;
DROP TABLE IF EXISTS fetch_runs;
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS vacancies;
DROP TABLE IF EXISTS users;
