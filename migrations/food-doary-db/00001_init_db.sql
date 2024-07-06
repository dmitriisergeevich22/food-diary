-- +goose Up
-- +goose StatementBegin

-- Таблица статусов пакета
CREATE TABLE public.package_status_enum (
    id SERIAL PRIMARY KEY,
    package_status_desc VARCHAR NOT NULL
);
INSERT INTO public.package_status_enum (package_status_desc) VALUES ('CREATED'), ('SUCCESS'), ('FAILED');

-- Таблица urls
CREATE TABLE public.destination_urls (
	id BIGSERIAL PRIMARY KEY,
	destination_url text UNIQUE
);

CREATE INDEX idx_destination_urls_url ON public.destination_urls (destination_url);

-- Таблица пакетов
CREATE TABLE public.packages (
	id BIGSERIAL PRIMARY KEY,
	type VARCHAR(40) NOT NULL,
    name VARCHAR(40) NOT NULL UNIQUE,
    destination_url_id INTEGER NOT NULL,
	receiver_is_hub BOOLEAN NOT NULL,
	receiver_operator_id VARCHAR(3) NOT NULL,
	sender_operator_id VARCHAR(3) NOT NULL,
    package_status_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	CONSTRAINT package_status_id_fkey FOREIGN KEY (package_status_id) REFERENCES public.package_status_enum(id),
	CONSTRAINT destination_url_id_fkey FOREIGN KEY (destination_url_id) REFERENCES public.destination_urls(id)
);

CREATE INDEX idx_packages_type ON public.packages (type);
CREATE INDEX idx_packages_name ON public.packages (name);
CREATE INDEX idx_packages_created_at ON public.packages (created_at);
CREATE INDEX idx_packages_updated_at ON public.packages (updated_at);

COMMENT ON COLUMN packages.type is 'Тип пакета (сообщения \ приглашение \ техн. квитанция).';
COMMENT ON COLUMN packages.name is 'Имя пакета (uuid4).';
COMMENT ON COLUMN packages.destination_url_id is 'Идентификатор URL для отправки пакета.';

-- Таблица ошибок обработки пакета
CREATE TABLE public.package_error (
    package_id BIGINT PRIMARY KEY,
	error_text text,
	error_code varchar(150),
	CONSTRAINT package_error_package_id_fkey FOREIGN KEY (package_id) REFERENCES public.packages(id)
);

CREATE INDEX idx_package_error_error_text ON public.package_error (error_text);
CREATE INDEX idx_package_error_error_code ON public.package_error (error_code);


-- Таблица типов события пакета
CREATE TABLE public.package_event_enum (
    id SERIAL PRIMARY KEY,
    package_event_desc VARCHAR NOT NULL
);
INSERT INTO public.package_event_enum (package_event_desc) VALUES ('PACKAGE_CREATED'), ('PACKAGE_GOT_AGAIN'), ('PACKAGE_REPROCESS'), ('PACKAGE_SUCCESS'), ('PACKAGE_SENT'), ('PACKAGE_ERROR');

-- Таблица событий пакета
CREATE TABLE public.package_events (
	id BIGSERIAL PRIMARY KEY,
	package_id BIGINT NOT NULL,
	package_event_id int NOT NULL,
	description text,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
	CONSTRAINT package_id_fkey FOREIGN KEY (package_id) REFERENCES public.packages(id),
	CONSTRAINT package_event_id_fkey FOREIGN KEY (package_event_id) REFERENCES public.package_event_enum(id)
);
CREATE INDEX idx_package_event_package_id ON public.package_events (package_id);
CREATE INDEX idx_package_event_created_at ON public.package_events (created_at);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

-- не будем сносить всю бд
SELECT 1;

-- +goose StatementEnd