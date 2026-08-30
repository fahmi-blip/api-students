CREATE TABLE IF NOT EXISTS students (
    id          SERIAL          PRIMARY KEY,
    nim         VARCHAR(20)     NOT NULL,
    name        VARCHAR(50)     NOT NULL,
    grade       NUMERIC(5,2)    NOT NULL,
    is_active   BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT Now()
);

CREATE UNIQUE INDEX IF NOT EXISTS students_nim_key
    ON students (nim);

CREATE UNIQUE INDEX IF NOT EXISTS students_name_lower_key
    ON students (LOWER(name));
