-- +goose Up
-- +goose StatementBegin

-- Таблица пользователей
CREATE TABLE users (
    id SERIAL PRIMARY KEY, -- Идентификатор
    created_at TIMESTAMPTZ DEFAULT NOW(), -- Дата создания
    name VARCHAR(100) NOT NULL, -- Имя 
    email VARCHAR(255) UNIQUE NOT NULL, -- Почта
    password VARCHAR(255) NOT NULL, -- Пароль
    phone VARCHAR(20) -- Телефон
);
CREATE INDEX idx_users_created ON users (created_at);

-- Таблица продуктов
CREATE TABLE products (
    id SERIAL PRIMARY KEY, -- Идентификатор
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Дата создания
	checked BOOLEAN DEFAULT FALSE, -- Проверен ли продукт
    name VARCHAR(255) NOT NULL, -- Название
    maker VARCHAR(255) NOT NULL, -- Производитель
    url VARCHAR(255), -- Ссылка на продукт в маркетплейсе
    logo VARCHAR(255), -- Логотип
    proteins FLOAT, -- Кол-во белков на 100г
    fats FLOAT, -- Кол-во жиров на 100г
    carbs FLOAT, -- Кол-во углеводов на 100г
    fiber FLOAT, -- Кол-во волокон на 100г
    calories FLOAT -- Кол-во калорий на 100г
);
CREATE INDEX idx_products_name ON products (name);
CREATE INDEX idx_products_maker ON products (maker);
CREATE INDEX idx_products_calories ON products (calories);

-- Таблица связи пользователей и продуктов
CREATE TABLE users2products (
    user_id INT NOT NULL,
    product_id INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, product_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);
CREATE INDEX idx_users2products_user_id ON users2products (user_id);
CREATE INDEX idx_users2products_product_id ON users2products (product_id);

-- 	Таблица событий
CREATE TABLE events (
    created_at TIMESTAMP NOT NULL,
    user_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    weight FLOAT NOT NULL,
    PRIMARY KEY (user_id, created_at)
);
CREATE INDEX idx_events_user_id ON events (user_id);
CREATE INDEX idx_events_product_id ON events (user_id);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

-- не будем сносить всю бд
SELECT 'БД откатана до последней миграции' as message;

-- +goose StatementEnd