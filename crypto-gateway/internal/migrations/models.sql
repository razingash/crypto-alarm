
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
    id BIGSERIAL,
    api character varying(500) NOT NULL,
    cooldown integer NOT NULL DEFAULT 20,
    is_actual boolean NOT NULL DEFAULT TRUE,
    is_history_on boolean NOT NULL DEFAULT FALSE,
    last_updated timestamp without time zone NOT NULL DEFAULT now()
);

CREATE TABLE crypto_api_history (
    id BIGSERIAL,
    crypto_api_id integer NOT NULL,
    weight smallint NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now()
);

CREATE TABLE crypto_currencies (
    id BIGSERIAL,
    currency character varying(100) NOT NULL,
    is_available boolean NOT NULL DEFAULT TRUE,
    last_updated timestamp without time zone NOT NULL DEFAULT now()
);

CREATE TABLE crypto_params (
    id BIGSERIAL,
    crypto_api_id integer NOT NULL,
    parameter character varying(500) NOT NULL,
    is_active boolean NOT NULL DEFAULT TRUE,
    last_updated timestamp without time zone NOT NULL DEFAULT now(),
    excluded_at timestamp without time zone
);

CREATE TABLE trigger_component (
    id BIGSERIAL,
    api_id integer NOT NULL,
    currency_id integer NOT NULL,
    parameter_id integer NOT NULL,
    name character varying(600) NOT NULL
);

CREATE TABLE trigger_component_history (
    id BIGSERIAL,
    trigger_history_id integer NOT NULL,
    component_id integer NOT NULL,
    value double precision NOT NULL
);

CREATE TABLE trigger_formula (
    id BIGSERIAL,
    formula character varying NOT NULL,
    formula_raw character varying NOT NULL,
    name character varying(150) NOT NULL,
    description character varying(1500),
    is_notified boolean NOT NULL DEFAULT FALSE,
    is_active boolean NOT NULL DEFAULT FALSE,
    is_history_on boolean NOT NULL DEFAULT FALSE,
    is_shutted_off boolean NOT NULL DEFAULT FALSE,
    last_triggered timestamp without time zone,
    cooldown integer NOT NULL DEFAULT 3600
);

CREATE TABLE trigger_formula_component (
    id BIGSERIAL,
    component_id integer NOT NULL,
    formula_id integer NOT NULL
);

CREATE TABLE trigger_history (
    id BIGSERIAL,
    formula_id integer NOT NULL,
    "timestamp" timestamp without time zone NOT NULL DEFAULT now(),
    status boolean NOT NULL
);

CREATE TABLE trigger_push_subscription (
    id BIGSERIAL,
    endpoint character varying NOT NULL,
    p256dh character varying NOT NULL,
    auth character varying NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now()
);
