create table public.coverage_areas (
  id uuid primary key default gen_random_uuid(),
  slug text not null unique,
  name text not null,
  geom extensions.geography(MultiPolygon, 4326),
  policies jsonb not null,
  enabled_sources text[] not null default '{}',
  active boolean not null default false,
  created_at timestamptz not null default now(),
  constraint coverage_areas_spot_ttl_positive check (
    (policies->>'spot_ttl_seconds')::integer > 0
  )
);

create table public.spots (
  id uuid primary key,
  cell_r10 bigint not null,
  location extensions.geography(Point, 4326) not null,
  status text not null default 'announced',
  confidence double precision not null,
  ttl_at timestamptz not null,
  releaser_id uuid not null references auth.users(id) on delete cascade,
  coverage_id uuid not null references public.coverage_areas(id),
  created_at timestamptz not null,
  constraint spots_status_valid check (status in ('announced', 'offered', 'claimed', 'occupied', 'expired', 'refuted')),
  constraint spots_confidence_valid check (confidence between 0 and 1),
  constraint spots_ttl_after_creation check (ttl_at > created_at)
);

create index spots_releaser_id_idx on public.spots (releaser_id);
create index spots_coverage_id_idx on public.spots (coverage_id);
create index spots_live_by_cell_idx on public.spots (cell_r10, ttl_at) where status in ('announced', 'offered', 'claimed');

create table public.outbox (
  id uuid primary key,
  type text not null,
  aggregate_id uuid not null,
  payload jsonb not null,
  occurred_at timestamptz not null,
  published_at timestamptz
);

create index outbox_aggregate_id_idx on public.outbox (aggregate_id);
create index outbox_unpublished_idx on public.outbox (occurred_at, id) where published_at is null;

create table public.timers (
  id uuid primary key default gen_random_uuid(),
  kind text not null,
  aggregate_id uuid not null,
  due_at timestamptz not null,
  payload jsonb not null default '{}',
  fired_at timestamptz,
  attempts integer not null default 0,
  created_at timestamptz not null default now(),
  constraint timers_kind_valid check (kind in ('offer_window', 'claim_expiry', 'spot_ttl', 'session_timeout')),
  constraint timers_attempts_nonnegative check (attempts >= 0)
);

create index timers_due_idx on public.timers (due_at, id) where fired_at is null;

alter table public.coverage_areas enable row level security;
alter table public.spots enable row level security;
alter table public.outbox enable row level security;
alter table public.timers enable row level security;

insert into public.coverage_areas (id, slug, name, policies, active)
values (
  '00000000-0000-4000-8000-000000000001',
  'madrid-region',
  'Comunidad de Madrid',
  '{"spot_ttl_seconds":180}'::jsonb,
  true
);
