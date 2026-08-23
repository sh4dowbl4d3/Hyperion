CREATE TABLE redirect_landing_stats (
    slug text PRIMARY KEY,
    hits integer NOT NULL DEFAULT 0
);

INSERT INTO redirect_landing_stats (slug, hits) VALUES
    ('promo', 0),
    ('newsletter', 0);
