-- UP
CREATE TABLE users
(
    id       serial PRIMARY KEY,
    username varchar UNIQUE NOT null,
    password bytea          NOT null
);

CREATE TABLE resources
(
    id      serial PRIMARY KEY,
    user_id int,
    type    int NOT null,
    data    bytea,
    meta    bytea,

    CONSTRAINT fk_users FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
-- DOWN
DROP TABLE IF EXISTS "resources";
DROP TABLE IF EXISTS "users";
