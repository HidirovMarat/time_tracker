CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    passport_number INT NOT NULL,
    passport_serie INT NOT NULL,
    surname VARCHAR(50),
    name VARCHAR(50),
    patronymic VARCHAR(50),
    address TEXT
);

CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    description TEXT NOT NULL,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id)
);
