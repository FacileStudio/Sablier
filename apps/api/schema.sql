
CREATE TABLE public.antenne_outbox (
    id bigint NOT NULL,
    channel text NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    idempotency_key text NOT NULL,
    sent_at timestamp with time zone
);

CREATE SEQUENCE public.antenne_outbox_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.antenne_outbox_id_seq OWNED BY public.antenne_outbox.id;

CREATE TABLE public.antenne_processed_events (
    idempotency_key text NOT NULL,
    processed_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.app_settings (
    id bigint NOT NULL,
    webhook_url text DEFAULT ''::text NOT NULL,
    webhook_secret_header text DEFAULT ''::text NOT NULL,
    webhook_secret_value text DEFAULT ''::text NOT NULL,
    antenne_url text DEFAULT ''::text NOT NULL,
    antenne_secret text DEFAULT ''::text NOT NULL,
    antenne_enabled boolean DEFAULT false NOT NULL,
    pool_events text DEFAULT ''::text NOT NULL
);

CREATE SEQUENCE public.app_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.app_settings_id_seq OWNED BY public.app_settings.id;

CREATE TABLE public.directory_deletion_state (
    id smallint NOT NULL,
    state bytea NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT directory_deletion_state_id_check CHECK ((id = 1))
);

CREATE TABLE public.draft_dismissals (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    started_at timestamp with time zone NOT NULL,
    stopped_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);

CREATE SEQUENCE public.draft_dismissals_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.draft_dismissals_id_seq OWNED BY public.draft_dismissals.id;

CREATE TABLE public.forfait_state (
    id bigint NOT NULL,
    document text NOT NULL,
    updated_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.heartbeats (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    at timestamp with time zone NOT NULL,
    source text NOT NULL,
    entity text DEFAULT ''::text NOT NULL,
    project_hint text DEFAULT ''::text NOT NULL,
    project_id bigint,
    write boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    registre_space_id text DEFAULT ''::text NOT NULL,
    opus_task_id text DEFAULT ''::text NOT NULL,
    title text DEFAULT ''::text NOT NULL
);

CREATE SEQUENCE public.heartbeats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.heartbeats_id_seq OWNED BY public.heartbeats.id;

CREATE TABLE public.mcp_actions (
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

CREATE TABLE public.porte_identities (
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

CREATE TABLE public.porte_login_codes (
    code_hash text NOT NULL,
    user_id bigint NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone,
    provider text DEFAULT ''::text NOT NULL,
    subject text DEFAULT ''::text NOT NULL
);

CREATE TABLE public.porte_sessions (
    id bigint NOT NULL,
    token_hash text NOT NULL,
    user_id bigint NOT NULL,
    label text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_used_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone
);

CREATE SEQUENCE public.porte_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.porte_sessions_id_seq OWNED BY public.porte_sessions.id;

CREATE TABLE public.projects (
    id bigint NOT NULL,
    name text,
    description text,
    icon text DEFAULT 'Layout'::text,
    owner_id bigint,
    space_id uuid,
    facile_id text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    synced_at timestamp with time zone
);

CREATE SEQUENCE public.projects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.projects_id_seq OWNED BY public.projects.id;

CREATE TABLE public.push_subscriptions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    endpoint text NOT NULL,
    p256dh text NOT NULL,
    auth text NOT NULL
);

CREATE SEQUENCE public.push_subscriptions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.push_subscriptions_id_seq OWNED BY public.push_subscriptions.id;

CREATE TABLE public.space_members (
    id text NOT NULL,
    space_id text NOT NULL,
    user_id bigint NOT NULL,
    role text DEFAULT 'member'::text NOT NULL,
    joined_at timestamp with time zone DEFAULT now()
);

CREATE TABLE public.spaces (
    id text NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    registre_space_id text
);

CREATE TABLE public.tasks (
    id bigint NOT NULL,
    project_id bigint,
    name text,
    status text DEFAULT 'to-do'::text,
    actor_id bigint,
    space_id uuid,
    facile_id text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    description text DEFAULT ''::text NOT NULL,
    due_date character varying(10),
    priority text DEFAULT 'none'::text NOT NULL,
    synced_name text DEFAULT ''::text NOT NULL,
    synced_at timestamp with time zone
);

CREATE SEQUENCE public.tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.tasks_id_seq OWNED BY public.tasks.id;

CREATE TABLE public.time_entries (
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
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    last_notification_at timestamp with time zone,
    cowork boolean DEFAULT false NOT NULL,
    attribution_source text,
    attribution_confidence double precision,
    needs_review boolean DEFAULT false NOT NULL
);

CREATE SEQUENCE public.time_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.time_entries_id_seq OWNED BY public.time_entries.id;

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text,
    name text,
    avatar_url text DEFAULT ''::text,
    avatar_source text DEFAULT ''::text,
    avatar_upload_path text DEFAULT ''::text,
    oidc_picture_url text,
    oidc_subject text,
    oidc_access_token text DEFAULT ''::text,
    oidc_refresh_token text DEFAULT ''::text,
    oidc_token_expiry timestamp with time zone DEFAULT '0001-01-01 00:00:00+00'::timestamp with time zone,
    profile_synced_at timestamp with time zone DEFAULT '0001-01-01 00:00:00+00'::timestamp with time zone,
    color text,
    password_hash text DEFAULT ''::text,
    rate numeric DEFAULT 0 NOT NULL,
    rate_type text DEFAULT 'daily'::text NOT NULL,
    workday_hours numeric DEFAULT 8 NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    auto_confirm_drafts boolean DEFAULT true NOT NULL,
    kind text DEFAULT 'human'::text NOT NULL,
    owner_user_id bigint,
    agent_key text,
    agent_avatar_url text DEFAULT ''::text NOT NULL,
    locale text DEFAULT 'en'::text NOT NULL,
    review_threshold double precision DEFAULT 0.8 NOT NULL
);

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

ALTER TABLE ONLY public.antenne_outbox ALTER COLUMN id SET DEFAULT nextval('public.antenne_outbox_id_seq'::regclass);

ALTER TABLE ONLY public.app_settings ALTER COLUMN id SET DEFAULT nextval('public.app_settings_id_seq'::regclass);

ALTER TABLE ONLY public.draft_dismissals ALTER COLUMN id SET DEFAULT nextval('public.draft_dismissals_id_seq'::regclass);

ALTER TABLE ONLY public.heartbeats ALTER COLUMN id SET DEFAULT nextval('public.heartbeats_id_seq'::regclass);

ALTER TABLE ONLY public.porte_sessions ALTER COLUMN id SET DEFAULT nextval('public.porte_sessions_id_seq'::regclass);

ALTER TABLE ONLY public.projects ALTER COLUMN id SET DEFAULT nextval('public.projects_id_seq'::regclass);

ALTER TABLE ONLY public.push_subscriptions ALTER COLUMN id SET DEFAULT nextval('public.push_subscriptions_id_seq'::regclass);

ALTER TABLE ONLY public.tasks ALTER COLUMN id SET DEFAULT nextval('public.tasks_id_seq'::regclass);

ALTER TABLE ONLY public.time_entries ALTER COLUMN id SET DEFAULT nextval('public.time_entries_id_seq'::regclass);

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);

ALTER TABLE ONLY public.antenne_outbox
    ADD CONSTRAINT antenne_outbox_idempotency_key_key UNIQUE (idempotency_key);

ALTER TABLE ONLY public.antenne_outbox
    ADD CONSTRAINT antenne_outbox_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.antenne_processed_events
    ADD CONSTRAINT antenne_processed_events_pkey PRIMARY KEY (idempotency_key);

ALTER TABLE ONLY public.app_settings
    ADD CONSTRAINT app_settings_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.directory_deletion_state
    ADD CONSTRAINT directory_deletion_state_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.draft_dismissals
    ADD CONSTRAINT draft_dismissals_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.forfait_state
    ADD CONSTRAINT forfait_state_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.heartbeats
    ADD CONSTRAINT heartbeats_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.mcp_actions
    ADD CONSTRAINT mcp_actions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.porte_identities
    ADD CONSTRAINT porte_identities_pkey PRIMARY KEY (provider, subject);

ALTER TABLE ONLY public.porte_login_codes
    ADD CONSTRAINT porte_login_codes_pkey PRIMARY KEY (code_hash);

ALTER TABLE ONLY public.porte_sessions
    ADD CONSTRAINT porte_sessions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.porte_sessions
    ADD CONSTRAINT porte_sessions_token_hash_key UNIQUE (token_hash);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.push_subscriptions
    ADD CONSTRAINT push_subscriptions_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.space_members
    ADD CONSTRAINT space_members_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.spaces
    ADD CONSTRAINT spaces_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.time_entries
    ADD CONSTRAINT time_entries_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE INDEX antenne_outbox_sent ON public.antenne_outbox USING btree (sent_at) WHERE (sent_at IS NOT NULL);

CREATE INDEX antenne_outbox_unsent ON public.antenne_outbox USING btree (id) WHERE (sent_at IS NULL);

CREATE INDEX idx_draft_dismissals_user_id ON public.draft_dismissals USING btree (user_id);

CREATE INDEX idx_heartbeats_at ON public.heartbeats USING btree (at);

CREATE UNIQUE INDEX idx_heartbeats_dedupe ON public.heartbeats USING btree (user_id, source, entity, at);

CREATE INDEX idx_heartbeats_user_at ON public.heartbeats USING btree (user_id, at);

CREATE INDEX idx_heartbeats_user_id ON public.heartbeats USING btree (user_id);

CREATE UNIQUE INDEX idx_projects_facile_id ON public.projects USING btree (facile_id);

CREATE INDEX idx_projects_owner_id ON public.projects USING btree (owner_id);

CREATE INDEX idx_projects_space_id ON public.projects USING btree (space_id);

CREATE UNIQUE INDEX idx_push_subscriptions_user_id ON public.push_subscriptions USING btree (user_id);

CREATE UNIQUE INDEX idx_space_user ON public.space_members USING btree (space_id, user_id);

CREATE UNIQUE INDEX idx_spaces_registre_space_id ON public.spaces USING btree (registre_space_id);

CREATE INDEX idx_tasks_actor_id ON public.tasks USING btree (actor_id);

CREATE UNIQUE INDEX idx_tasks_facile_id ON public.tasks USING btree (facile_id);

CREATE INDEX idx_tasks_project_id ON public.tasks USING btree (project_id);

CREATE UNIQUE INDEX idx_tasks_project_name ON public.tasks USING btree (project_id, name);

CREATE INDEX idx_tasks_space_id ON public.tasks USING btree (space_id);

CREATE UNIQUE INDEX idx_time_entries_facile_id ON public.time_entries USING btree (facile_id);

CREATE INDEX idx_time_entries_needs_review ON public.time_entries USING btree (needs_review);

CREATE INDEX idx_time_entries_project_id ON public.time_entries USING btree (project_id);

CREATE INDEX idx_time_entries_space_id ON public.time_entries USING btree (space_id);

CREATE INDEX idx_time_entries_task_id ON public.time_entries USING btree (task_id);

CREATE INDEX idx_time_entries_user_id ON public.time_entries USING btree (user_id);

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);

CREATE UNIQUE INDEX idx_users_o_id_c_subject ON public.users USING btree (oidc_subject);

CREATE UNIQUE INDEX idx_users_owner_agent_key ON public.users USING btree (owner_user_id, agent_key);

CREATE INDEX idx_users_owner_user_id ON public.users USING btree (owner_user_id);

CREATE INDEX mcp_actions_open_idx ON public.mcp_actions USING btree (user_id, created_at DESC) WHERE (resolved_at IS NULL);

CREATE INDEX porte_identities_user_idx ON public.porte_identities USING btree (user_id);

CREATE INDEX porte_sessions_expiry_idx ON public.porte_sessions USING btree (expires_at) WHERE (expires_at IS NOT NULL);

CREATE INDEX porte_sessions_user_idx ON public.porte_sessions USING btree (user_id);

ALTER TABLE ONLY public.porte_identities
    ADD CONSTRAINT porte_identities_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.porte_login_codes
    ADD CONSTRAINT porte_login_codes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.porte_sessions
    ADD CONSTRAINT porte_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

