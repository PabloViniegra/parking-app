update public.coverage_areas set active = false where geom is null;

update public.outbox set payload = payload #- '{data,releaser_id}'
where type = 'availability.spot.announced.v1';

alter table public.coverage_areas add constraint coverage_areas_active_requires_geometry
  check (not active or (geom is not null and not extensions.st_isempty(geom::extensions.geometry) and extensions.st_isvalid(geom::extensions.geometry)));

create table public.spot_daily_counts (
  cell_r10 bigint not null,
  day date not null,
  announcements bigint not null check (announcements >= 0),
  primary key (cell_r10, day)
);

alter table public.spot_daily_counts enable row level security;
create index spots_retention_idx on public.spots (created_at);
create index outbox_retention_idx on public.outbox (occurred_at);

create schema if not exists private;
revoke all on schema private from public, anon, authenticated;

create function private.purge_spot_locations(cutoff timestamptz)
returns void
language sql
security invoker
set search_path = ''
as $$
  with removed as (
    delete from public.spots
    where created_at <= cutoff - interval '23 hours'
    returning cell_r10, created_at
  )
  insert into public.spot_daily_counts (cell_r10, day, announcements)
  select cell_r10, (created_at at time zone 'UTC')::date, count(*)
  from removed
  group by cell_r10, (created_at at time zone 'UTC')::date
  on conflict (cell_r10, day) do update
  set announcements = public.spot_daily_counts.announcements + excluded.announcements;

  delete from public.outbox where occurred_at <= cutoff - interval '7 days';
$$;

revoke all on function private.purge_spot_locations(timestamptz) from public, anon, authenticated;

-- Purging at 23 hours leaves room for the hourly schedule within the 24-hour privacy limit (ADR-0010).
select cron.schedule('purge-spot-locations', '0 * * * *', $$select private.purge_spot_locations(now())$$);
