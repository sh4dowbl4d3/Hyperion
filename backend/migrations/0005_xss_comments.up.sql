CREATE TABLE xss_comments (
    id serial PRIMARY KEY,
    author text NOT NULL,
    body text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO xss_comments (author, body) VALUES
    ('moderator', 'Welcome! Keep feedback constructive.'),
    ('priya', 'Great write-up, the diagrams helped a lot.');
