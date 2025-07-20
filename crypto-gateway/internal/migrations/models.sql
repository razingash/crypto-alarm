
SET statement_timeout = 0;
SET lock_timeout = 0;
--SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
--SET xmloption = content;
--SET client_min_messages = warning;
SET row_security = off; -- скорее всего не нужно будет включать
SET default_tablespace = '';
SET default_table_access_method = heap;
SET search_path TO public;

CREATE TABLE crypto_api (
    id BIGSERIAL PRIMARY KEY,
    api character varying(500) NOT NULL,
    cooldown integer NOT NULL DEFAULT 20,
    is_actual boolean NOT NULL DEFAULT TRUE,
    is_history_on boolean NOT NULL DEFAULT FALSE,
    is_accessible boolean NOT NULL DEFAULT TRUE,
    last_updated timestamp without time zone NOT NULL DEFAULT now()
);

CREATE TABLE crypto_api_history (
    id BIGSERIAL PRIMARY KEY,
    crypto_api_id integer NOT NULL REFERENCES crypto_api(id) ON DELETE CASCADE,
    weight smallint NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now()
);

CREATE TABLE crypto_currencies (
    id BIGSERIAL PRIMARY KEY,
    currency character varying(100) NOT NULL,
    is_available boolean NOT NULL DEFAULT TRUE,
    last_updated timestamp without time zone NOT NULL DEFAULT now()
);

CREATE TABLE trigger_formula (
    id BIGSERIAL PRIMARY KEY,
    formula character varying NOT NULL,
    formula_raw character varying NOT NULL
);

CREATE TABLE strategy_history ( -- адаптировать под прогон - чтобы тут были колебания от общего депозита, сейчас от неё нет смысла
    id BIGSERIAL PRIMARY KEY,
    formula_id integer NOT NULL REFERENCES trigger_formula(id) ON DELETE CASCADE,
    "timestamp" timestamp without time zone NOT NULL DEFAULT now(),
    status boolean NOT NULL
);

CREATE TABLE crypto_params (
    id BIGSERIAL PRIMARY KEY,
    crypto_api_id integer NOT NULL REFERENCES crypto_api(id) ON DELETE CASCADE,
    parameter character varying(500) NOT NULL,
    is_active boolean NOT NULL DEFAULT TRUE,
    last_updated timestamp without time zone NOT NULL DEFAULT now(),
    excluded_at timestamp without time zone
);

CREATE TABLE trigger_component (
    id BIGSERIAL PRIMARY KEY,
    api_id integer NOT NULL REFERENCES crypto_api(id) ON DELETE CASCADE,
    currency_id integer NOT NULL REFERENCES crypto_currencies(id) ON DELETE CASCADE,
    parameter_id integer NOT NULL REFERENCES crypto_params(id) ON DELETE CASCADE,
    name character varying(600) NOT NULL
);

CREATE TABLE trigger_component_history (
    id BIGSERIAL PRIMARY KEY,
    expression_id integer NOT NULL REFERENCES strategy_history(id) ON DELETE CASCADE,
    component_id integer NOT NULL REFERENCES trigger_component(id) ON DELETE CASCADE,
    value double precision NOT NULL
);

CREATE TABLE crypto_strategy (
    id BIGSERIAL PRIMARY KEY,
    name character varying(150) NOT NULL,
    description character varying(1500),
    is_notified boolean NOT NULL DEFAULT FALSE,
    is_active boolean NOT NULL DEFAULT FALSE,
    is_history_on boolean NOT NULL DEFAULT FALSE,
    is_shutted_off boolean NOT NULL DEFAULT FALSE,
    last_triggered timestamp without time zone,
    cooldown integer NOT NULL DEFAULT 3600
);

CREATE TABLE crypto_strategy_formula (
    id BIGSERIAL PRIMARY KEY,
    strategy_id integer NOT NULL REFERENCES crypto_strategy(id) ON DELETE CASCADE,
    formula_id integer NOT NULL REFERENCES trigger_formula(id) ON DELETE CASCADE
);

CREATE TABLE trigger_formula_component (
    id BIGSERIAL PRIMARY KEY,
    component_id integer NOT NULL REFERENCES trigger_component(id) ON DELETE CASCADE,
    formula_id integer NOT NULL REFERENCES trigger_formula(id) ON DELETE CASCADE
);

CREATE TABLE trigger_push_subscription (
    id BIGSERIAL PRIMARY KEY,
    endpoint character varying NOT NULL,
    p256dh character varying NOT NULL,
    auth character varying NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now()
);

CREATE TABLE settings (
    id BIGSERIAL PRIMARY KEY,
    name character varying(150) NOT NULL,
    is_active boolean NOT NULL DEFAULT TRUE
);