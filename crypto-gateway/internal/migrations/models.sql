
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
    api varchar(500) NOT NULL,
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
    currency varchar(100) NOT NULL,
    is_available boolean NOT NULL DEFAULT TRUE,
    last_updated timestamp without time zone NOT NULL DEFAULT now()
);
CREATE TABLE crypto_params (
    id BIGSERIAL PRIMARY KEY,
    crypto_api_id integer NOT NULL REFERENCES crypto_api(id) ON DELETE CASCADE,
    parameter varchar(500) NOT NULL,
    is_active boolean NOT NULL DEFAULT TRUE,
    last_updated timestamp without time zone NOT NULL DEFAULT now(),
    excluded_at timestamp without time zone
);

CREATE TABLE crypto_variables (
    id BIGSERIAL PRIMARY KEY,
    symbol varchar(40) NOT NULL UNIQUE,
    name varchar(255) NOT NULL,
    description varchar,
    formula varchar NOT NULL,
    formula_raw varchar NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);
CREATE TABLE crypto_variables_api (
    id BIGSERIAL PRIMARY KEY,
    api_id integer NOT NULL REFERENCES crypto_api(id) ON DELETE CASCADE,
    variable_id integer NOT NULL REFERENCES crypto_variables(id) ON DELETE CASCADE,
    parameter_id integer NOT NULL REFERENCES crypto_params(id) ON DELETE CASCADE
);

-- переменная может хранить в себе другие переменные.(пока нет)
CREATE TABLE crypto_variable_links (
    id BIGSERIAL PRIMARY KEY,
    source_variable_id BIGINT NOT NULL REFERENCES crypto_variables(id) ON DELETE CASCADE,
    target_variable_id BIGINT NOT NULL REFERENCES crypto_variables(id) ON DELETE CASCADE,
    CONSTRAINT no_self_reference CHECK (source_variable_id <> target_variable_id),
    UNIQUE (source_variable_id, target_variable_id)
);

CREATE TABLE trigger_formula (
    id BIGSERIAL PRIMARY KEY,
    formula varchar NOT NULL,
    formula_raw varchar NOT NULL
);

CREATE TABLE strategy_history ( -- адаптировать под прогон - чтобы тут были колебания от общего депозита, сейчас от неё нет смысла
    id BIGSERIAL PRIMARY KEY,
    formula_id integer NOT NULL REFERENCES trigger_formula(id) ON DELETE CASCADE,
    "timestamp" timestamp without time zone NOT NULL DEFAULT now(),
    status boolean NOT NULL
);

CREATE TABLE trigger_component (
    id BIGSERIAL PRIMARY KEY,
    api_id integer NOT NULL REFERENCES crypto_api(id) ON DELETE CASCADE,
    currency_id integer NOT NULL REFERENCES crypto_currencies(id) ON DELETE CASCADE,
    parameter_id integer NOT NULL REFERENCES crypto_params(id) ON DELETE CASCADE,
    name varchar(600) NOT NULL
);

CREATE TABLE trigger_component_history (
    id BIGSERIAL PRIMARY KEY,
    expression_id integer NOT NULL REFERENCES strategy_history(id) ON DELETE CASCADE,
    component_id integer NOT NULL REFERENCES trigger_component(id) ON DELETE CASCADE,
    value double precision NOT NULL
);

CREATE TABLE crypto_strategy (
    id BIGSERIAL PRIMARY KEY,
    name varchar(150) NOT NULL,
    description varchar(1500),
    is_notified boolean NOT NULL DEFAULT FALSE,
    is_active boolean NOT NULL DEFAULT FALSE,
    is_history_on boolean NOT NULL DEFAULT FALSE,
    is_shutted_off boolean NOT NULL DEFAULT FALSE,
    last_triggered timestamp without time zone,
    cooldown integer NOT NULL DEFAULT 3600
);

-- универсальный оркестратор
CREATE TABLE module_orchestrator (
    id BIGSERIAL PRIMARY KEY,
    --name VARCHAR(150),
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- предполагается, что для конкретного сигнала source_type должен быть один, если нужно учитывать данные сразу с
-- двух источников - например Binance и Bybit, тогда нужно будет делать систему из оркестраторов
CREATE TABLE orchestrator_inputs (
    id BIGSERIAL PRIMARY KEY,
    orchestrator_id BIGINT NOT NULL REFERENCES module_orchestrator(id) ON DELETE CASCADE,
    formula TEXT NOT NULL,
    tag VARCHAR(100) NOT NULL
);

CREATE TABLE orchestrator_input_sources (
    id BIGSERIAL PRIMARY KEY,
    input_id BIGINT NOT NULL REFERENCES orchestrator_inputs(id) ON DELETE CASCADE,
    source_type VARCHAR(100) NOT NULL,  -- 'binance', 'nasdaq', 'custom'| определяет из какой таблицы брать source_id
    source_id BIGINT NOT NULL -- является ссылкой на id в необходимой таблице
);

CREATE TABLE crypto_strategy_variable (
    id BIGSERIAL PRIMARY KEY,
    strategy_id integer NOT NULL REFERENCES crypto_strategy(id) ON DELETE CASCADE,
    crypto_variable_id integer NOT NULL REFERENCES crypto_variables(id) ON DELETE CASCADE,
    crypto_currency_id integer NOT NULL REFERENCES crypto_currencies(id) ON DELETE CASCADE
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
    endpoint varchar NOT NULL,
    p256dh varchar NOT NULL,
    auth varchar NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now()
);

CREATE TABLE diagrams (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    data JSON,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE settings (
    id BIGSERIAL PRIMARY KEY,
    name varchar(150) NOT NULL,
    is_active boolean NOT NULL DEFAULT TRUE
);