-- Миграция: add_users_table
-- Создано: 2025-10-08 19:13:02

CREATE TABLE IF NOT EXISTS users (
    id UInt32,
    name String,
    email String,
    password_hash String DEFAULT '',
    role String DEFAULT 'student',
    is_active UInt8 DEFAULT 1,
    last_login Nullable(DateTime),
    created_at DateTime DEFAULT now(),
    updated_at DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY id;
