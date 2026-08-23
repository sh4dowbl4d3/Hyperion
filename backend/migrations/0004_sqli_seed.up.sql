CREATE TABLE sqli_contacts (
    id serial PRIMARY KEY,
    name text NOT NULL,
    email text NOT NULL,
    department text NOT NULL,
    note text NOT NULL DEFAULT '',
    hidden boolean NOT NULL DEFAULT false
);

INSERT INTO sqli_contacts (name, email, department, note, hidden) VALUES
    ('Alice Nguyen',  'alice@example.test',  'Engineering', 'Prefers async updates.', false),
    ('Bob Martinez',  'bob@example.test',    'Engineering', 'On call odd weeks.',     false),
    ('Carla Singh',   'carla@example.test',  'Security',    'Reviews PRs daily.',     false),
    ('Dmitri Ivanov', 'dmitri@example.test', 'Sales',       'Travels most Tuesdays.', false),
    ('Elena Rossi',   'elena@example.test',  'Support',     'Owns the FAQ page.',     false),
    ('Archive Vault', 'vault@example.test',  'REDACTED',    'FLAG-SQLI-9d41c2',       true);
