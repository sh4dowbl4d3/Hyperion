CREATE TABLE idor_documents (
    id serial PRIMARY KEY,
    owner_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title text NOT NULL,
    classification text NOT NULL DEFAULT 'internal',
    content text NOT NULL
);

CREATE INDEX idx_idor_documents_owner ON idor_documents (owner_id);
