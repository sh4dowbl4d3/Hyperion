CREATE TABLE race_coupons (
    code text PRIMARY KEY,
    max_redemptions integer NOT NULL CHECK (max_redemptions > 0),
    redeemed integer NOT NULL DEFAULT 0
);

CREATE TABLE race_redemptions (
    id serial PRIMARY KEY,
    code text NOT NULL REFERENCES race_coupons(code) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO race_coupons (code, max_redemptions, redeemed) VALUES
    ('LAUNCH-2026', 5, 0);
