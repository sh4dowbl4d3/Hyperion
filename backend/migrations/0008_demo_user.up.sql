-- Shared demo account so reviewers can sign in without registering.
-- Email:    demo@hyperion.test
-- Password: Password123!
INSERT INTO users (email, password_hash, role_id)
SELECT 'demo@hyperion.test',
       '$2a$12$517uz62PTMHtSD1RCS0eo.mFuY12XXSGDGHHLHDbmi8o4qS1yKqja',
       1
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE email = 'demo@hyperion.test'
);
