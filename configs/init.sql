CREATE TABLE IF NOT EXISTS requests (
    id SERIAL NOT NULL PRIMARY KEY,
    method TEXT NOT NULL,
    host TEXT NOT NULL,
    scheme TEXT NOT NULL,
    path TEXT NOT NULL,
    headers TEXT NOT NULL,
    body TEXT NOT NULL,
    params TEXT NOT NULL
);
