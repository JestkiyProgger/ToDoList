CREATE SCHEMA ToDoList;

CREATE TABLE ToDoList.users (
    id SERIAL PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 1,
    name VARCHAR(50) NOT NULL CHECK (CHAR_LENGTH(name) BETWEEN 3 AND 50),
    phone_number VARCHAR(15) CHECK (
        phone_number ~ '^\+[0-9]+$'
        AND
        CHAR_LENGTH(phone_number) BETWEEN 10 AND 15
    )
);

CREATE TABLE ToDoList.tasks (
    id SERIAL PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 1,
    title VARCHAR(100) NOT NULL CHECK (CHAR_LENGTH(title) BETWEEN 1 AND 100),
    description VARCHAR(1000) CHECK (CHAR_LENGTH(title) BETWEEN 1 AND 1000),
    completed BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,

    CHECK (
        (completed=FALSE AND completed_at IS NULL)
        OR
        (completed=TRUE AND completed_at IS NOT NULL AND completed_at >= created_at)
    ),

    user_id INTEGER NOT NULL REFERENCES ToDoList.users(id)
);