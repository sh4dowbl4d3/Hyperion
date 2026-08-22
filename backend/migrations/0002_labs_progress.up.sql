CREATE TABLE labs (
    slug text PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    category text NOT NULL,
    difficulty text NOT NULL,
    xp integer NOT NULL CHECK (xp >= 0),
    vulnerability_type text NOT NULL,
    objective text NOT NULL DEFAULT ''
);

CREATE TABLE progress (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lab_slug text NOT NULL REFERENCES labs(slug),
    status text NOT NULL DEFAULT 'not_started',
    xp_awarded integer NOT NULL DEFAULT 0 CHECK (xp_awarded >= 0),
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, lab_slug),
    CONSTRAINT progress_status_valid CHECK (status IN ('not_started', 'in_progress', 'completed')),
    CONSTRAINT progress_completed_requires_xp CHECK (
        (status = 'completed' AND xp_awarded > 0 AND completed_at IS NOT NULL)
        OR (status <> 'completed' AND completed_at IS NULL)
    )
);
