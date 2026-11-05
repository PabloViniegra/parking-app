create table public.spot_announcements (
  releaser_id uuid not null references auth.users(id) on delete cascade,
  idempotency_key uuid not null,
  fingerprint text not null,
  spot_id uuid not null references public.spots(id) on delete cascade deferrable initially deferred,
  primary key (releaser_id, idempotency_key)
);

create index spot_announcements_spot_id_idx on public.spot_announcements (spot_id);
alter table public.spot_announcements enable row level security;
