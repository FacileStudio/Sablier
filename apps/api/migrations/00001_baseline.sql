-- +goose Up
-- baseline: the production schema as GORM AutoMigrate and schemas.Migrate built it, captured 2026-11-13 with pg_dump --schema-only --no-owner --no-privileges

CREATE TABLE antenne_outbox (
    id bigint NOT NULL,
    channel text NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    idempotency_key text NOT NULL,
    sent_at timestamp with time zone
);

CREATE SEQUENCE antenne_outbox_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE antenne_outbox_id_seq OWNED BY antenne_outbox.id;

CREATE TABLE antenne_processed_events (
    idempotency_key text NOT NULL,
    processed_at timestamp with time zone
);

CREATE TABLE app_settings (
    id bigint NOT NULL,
    webhook_url text DEFAULT ''::text NOT NULL,
    webhook_secret_header text DEFAULT ''::text NOT NULL,
    webhook_secret_value text DEFAULT ''::text NOT NULL,
    antenne_url text DEFAULT ''::text NOT NULL,
    antenne_secret text DEFAULT ''::text NOT NULL,
    antenne_enabled boolean DEFAULT false NOT NULL,
    pool_events text DEFAULT ''::text NOT NULL
);

CREATE SEQUENCE app_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE app_settings_id_seq OWNED BY app_settings.id;

CREATE TABLE directory_deletion_state (
    id smallint NOT NULL,
    state bytea NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT directory_deletion_state_id_check CHECK ((id = 1))
);

CREATE TABLE draft_dismissals (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    started_at timestamp with time zone NOT NULL,
    stopped_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone
);

CREATE SEQUENCE draft_dismissals_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE draft_dismissals_id_seq OWNED BY draft_dismissals.id;

CREATE TABLE forfait_state (
    id bigint NOT NULL,
    document text NOT NULL,
    updated_at timestamp with time zone
);

CREATE TABLE heartbeats (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    at timestamp with time zone NOT NULL,
    source text NOT NULL,
    entity text DEFAULT ''::text NOT NULL,
    project_hint text DEFAULT ''::text NOT NULL,
    project_id bigint,
    write boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone,
    registre_space_id text DEFAULT ''::text NOT NULL,
    opus_task_id text DEFAULT ''::text NOT NULL,
    title text DEFAULT ''::text NOT NULL
);

CREATE SEQUENCE heartbeats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE heartbeats_id_seq OWNED BY heartbeats.id;

CREATE TABLE mcp_actions (
    id text NOT NULL,
    user_id text NOT NULL,
    kind text NOT NULL,
    target text NOT NULL,
    digest text NOT NULL,
    summary text DEFAULT ''::text NOT NULL,
    client_id text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    resolved_at timestamp with time zone,
    outcome text DEFAULT ''::text NOT NULL
);

CREATE TABLE porte_identities (
    user_id bigint NOT NULL,
    provider text NOT NULL,
    subject text NOT NULL,
    password_hash text DEFAULT ''::text NOT NULL,
    access_token text DEFAULT ''::text NOT NULL,
    refresh_token text DEFAULT ''::text NOT NULL,
    token_expiry timestamp with time zone,
    roles jsonb,
    roles_synced_at timestamp with time zone,
    synced_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);

CREATE TABLE porte_login_codes (
    code_hash text NOT NULL,
    user_id bigint NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone,
    provider text DEFAULT ''::text NOT NULL,
    subject text DEFAULT ''::text NOT NULL
);

CREATE TABLE porte_sessions (
    id bigint NOT NULL,
    token_hash text NOT NULL,
    user_id bigint NOT NULL,
    label text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_used_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone
);

CREATE SEQUENCE porte_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE porte_sessions_id_seq OWNED BY porte_sessions.id;

CREATE TABLE projects (
    id bigint NOT NULL,
    name text,
    description text,
    icon text DEFAULT 'Layout'::text,
    owner_id bigint,
    space_id uuid,
    facile_id text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    synced_at timestamp with time zone
);

CREATE SEQUENCE projects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE projects_id_seq OWNED BY projects.id;

CREATE TABLE push_subscriptions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    endpoint text NOT NULL,
    p256dh text NOT NULL,
    auth text NOT NULL
);

CREATE SEQUENCE push_subscriptions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE push_subscriptions_id_seq OWNED BY push_subscriptions.id;

CREATE TABLE space_members (
    id text NOT NULL,
    space_id text NOT NULL,
    user_id bigint NOT NULL,
    role text DEFAULT 'member'::text NOT NULL,
    joined_at timestamp with time zone
);

CREATE TABLE spaces (
    id text NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    registre_space_id text
);

CREATE TABLE tasks (
    id bigint NOT NULL,
    project_id bigint,
    name text,
    status text DEFAULT 'to-do'::text,
    actor_id bigint,
    space_id uuid,
    facile_id text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    description text DEFAULT ''::text NOT NULL,
    due_date character varying(10),
    priority text DEFAULT 'none'::text NOT NULL,
    synced_name text DEFAULT ''::text NOT NULL,
    synced_at timestamp with time zone
);

CREATE SEQUENCE tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE tasks_id_seq OWNED BY tasks.id;

CREATE TABLE time_entries (
    id bigint NOT NULL,
    project_id bigint,
    task_id bigint,
    user_id bigint,
    space_id uuid,
    facile_id text,
    description text,
    started_at timestamp with time zone,
    stopped_at timestamp with time zone,
    paused_at timestamp with time zone,
    paused_duration_ms bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    last_notification_at timestamp with time zone,
    cowork boolean DEFAULT false NOT NULL,
    attribution_source text,
    attribution_confidence double precision,
    needs_review boolean DEFAULT false NOT NULL
);

CREATE SEQUENCE time_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE time_entries_id_seq OWNED BY time_entries.id;

CREATE TABLE users (
    id bigint NOT NULL,
    email text,
    name text,
    avatar_url text,
    avatar_source text,
    avatar_upload_path text,
    oidc_picture_url text,
    oidc_subject text,
    oidc_access_token text,
    oidc_refresh_token text,
    oidc_token_expiry timestamp with time zone,
    profile_synced_at timestamp with time zone,
    color text,
    password_hash text,
    rate numeric DEFAULT 0 NOT NULL,
    rate_type text DEFAULT 'daily'::text NOT NULL,
    workday_hours numeric DEFAULT 8 NOT NULL,
    created_at timestamp with time zone,
    auto_confirm_drafts boolean DEFAULT true NOT NULL,
    kind text DEFAULT 'human'::text NOT NULL,
    owner_user_id bigint,
    agent_key text,
    agent_avatar_url text DEFAULT ''::text NOT NULL,
    locale text DEFAULT 'en'::text NOT NULL,
    review_threshold double precision DEFAULT 0.8 NOT NULL
);

CREATE SEQUENCE users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE users_id_seq OWNED BY users.id;

ALTER TABLE ONLY antenne_outbox ALTER COLUMN id SET DEFAULT nextval('antenne_outbox_id_seq'::regclass);

ALTER TABLE ONLY app_settings ALTER COLUMN id SET DEFAULT nextval('app_settings_id_seq'::regclass);

ALTER TABLE ONLY draft_dismissals ALTER COLUMN id SET DEFAULT nextval('draft_dismissals_id_seq'::regclass);

ALTER TABLE ONLY heartbeats ALTER COLUMN id SET DEFAULT nextval('heartbeats_id_seq'::regclass);

ALTER TABLE ONLY porte_sessions ALTER COLUMN id SET DEFAULT nextval('porte_sessions_id_seq'::regclass);

ALTER TABLE ONLY projects ALTER COLUMN id SET DEFAULT nextval('projects_id_seq'::regclass);

ALTER TABLE ONLY push_subscriptions ALTER COLUMN id SET DEFAULT nextval('push_subscriptions_id_seq'::regclass);

ALTER TABLE ONLY tasks ALTER COLUMN id SET DEFAULT nextval('tasks_id_seq'::regclass);

ALTER TABLE ONLY time_entries ALTER COLUMN id SET DEFAULT nextval('time_entries_id_seq'::regclass);

ALTER TABLE ONLY users ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);

ALTER TABLE ONLY antenne_outbox
    ADD CONSTRAINT antenne_outbox_idempotency_key_key UNIQUE (idempotency_key);

ALTER TABLE ONLY antenne_outbox
    ADD CONSTRAINT antenne_outbox_pkey PRIMARY KEY (id);

ALTER TABLE ONLY antenne_processed_events
    ADD CONSTRAINT antenne_processed_events_pkey PRIMARY KEY (idempotency_key);

ALTER TABLE ONLY app_settings
    ADD CONSTRAINT app_settings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY directory_deletion_state
    ADD CONSTRAINT directory_deletion_state_pkey PRIMARY KEY (id);

ALTER TABLE ONLY draft_dismissals
    ADD CONSTRAINT draft_dismissals_pkey PRIMARY KEY (id);

ALTER TABLE ONLY forfait_state
    ADD CONSTRAINT forfait_state_pkey PRIMARY KEY (id);

ALTER TABLE ONLY heartbeats
    ADD CONSTRAINT heartbeats_pkey PRIMARY KEY (id);

ALTER TABLE ONLY mcp_actions
    ADD CONSTRAINT mcp_actions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY porte_identities
    ADD CONSTRAINT porte_identities_pkey PRIMARY KEY (provider, subject);

ALTER TABLE ONLY porte_login_codes
    ADD CONSTRAINT porte_login_codes_pkey PRIMARY KEY (code_hash);

ALTER TABLE ONLY porte_sessions
    ADD CONSTRAINT porte_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY porte_sessions
    ADD CONSTRAINT porte_sessions_token_hash_key UNIQUE (token_hash);

ALTER TABLE ONLY projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);

ALTER TABLE ONLY push_subscriptions
    ADD CONSTRAINT push_subscriptions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY space_members
    ADD CONSTRAINT space_members_pkey PRIMARY KEY (id);

ALTER TABLE ONLY spaces
    ADD CONSTRAINT spaces_pkey PRIMARY KEY (id);

ALTER TABLE ONLY tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY time_entries
    ADD CONSTRAINT time_entries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE INDEX antenne_outbox_sent ON antenne_outbox USING btree (sent_at) WHERE (sent_at IS NOT NULL);

CREATE INDEX antenne_outbox_unsent ON antenne_outbox USING btree (id) WHERE (sent_at IS NULL);

CREATE INDEX idx_draft_dismissals_user_id ON draft_dismissals USING btree (user_id);

CREATE INDEX idx_heartbeats_at ON heartbeats USING btree (at);

CREATE UNIQUE INDEX idx_heartbeats_dedupe ON heartbeats USING btree (user_id, source, entity, at);

CREATE INDEX idx_heartbeats_user_at ON heartbeats USING btree (user_id, at);

CREATE INDEX idx_heartbeats_user_id ON heartbeats USING btree (user_id);

CREATE UNIQUE INDEX idx_projects_facile_id ON projects USING btree (facile_id);

CREATE INDEX idx_projects_owner_id ON projects USING btree (owner_id);

CREATE INDEX idx_projects_space_id ON projects USING btree (space_id);

CREATE UNIQUE INDEX idx_push_subscriptions_user_id ON push_subscriptions USING btree (user_id);

CREATE UNIQUE INDEX idx_space_user ON space_members USING btree (space_id, user_id);

CREATE UNIQUE INDEX idx_spaces_registre_space_id ON spaces USING btree (registre_space_id);

CREATE INDEX idx_tasks_actor_id ON tasks USING btree (actor_id);

CREATE UNIQUE INDEX idx_tasks_facile_id ON tasks USING btree (facile_id);

CREATE INDEX idx_tasks_project_id ON tasks USING btree (project_id);

CREATE UNIQUE INDEX idx_tasks_project_name ON tasks USING btree (project_id, name);

CREATE INDEX idx_tasks_space_id ON tasks USING btree (space_id);

CREATE UNIQUE INDEX idx_time_entries_facile_id ON time_entries USING btree (facile_id);

CREATE INDEX idx_time_entries_needs_review ON time_entries USING btree (needs_review);

CREATE INDEX idx_time_entries_project_id ON time_entries USING btree (project_id);

CREATE INDEX idx_time_entries_space_id ON time_entries USING btree (space_id);

CREATE INDEX idx_time_entries_task_id ON time_entries USING btree (task_id);

CREATE INDEX idx_time_entries_user_id ON time_entries USING btree (user_id);

CREATE UNIQUE INDEX idx_users_email ON users USING btree (email);

CREATE UNIQUE INDEX idx_users_o_id_c_subject ON users USING btree (oidc_subject);

CREATE UNIQUE INDEX idx_users_owner_agent_key ON users USING btree (owner_user_id, agent_key);

CREATE INDEX idx_users_owner_user_id ON users USING btree (owner_user_id);

CREATE INDEX mcp_actions_open_idx ON mcp_actions USING btree (user_id, created_at DESC) WHERE (resolved_at IS NULL);

CREATE INDEX porte_identities_user_idx ON porte_identities USING btree (user_id);

CREATE INDEX porte_sessions_expiry_idx ON porte_sessions USING btree (expires_at) WHERE (expires_at IS NOT NULL);

CREATE INDEX porte_sessions_user_idx ON porte_sessions USING btree (user_id);

ALTER TABLE ONLY porte_identities
    ADD CONSTRAINT porte_identities_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY porte_login_codes
    ADD CONSTRAINT porte_login_codes_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE ONLY porte_sessions
    ADD CONSTRAINT porte_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
SELECT 'baseline migration is not reversible' WHERE false;
