do $$
begin
  if not exists (select from pg_roles where rolname = 'parking_core') then
    create role parking_core nologin noinherit nosuperuser nocreatedb nocreaterole noreplication nobypassrls;
  end if;
end;
$$;

grant usage on schema public, extensions to parking_core;
grant select (id, slug, geom, policies, active) on public.coverage_areas to parking_core;
grant select (id, ttl_at) on public.spots to parking_core;
grant insert (id, cell_r10, location, status, confidence, ttl_at, releaser_id, coverage_id, created_at)
  on public.spots to parking_core;
grant select, insert on public.spot_announcements to parking_core;
grant insert (id, type, aggregate_id, payload, occurred_at) on public.outbox to parking_core;

create policy core_reads_active_coverage on public.coverage_areas
  for select to parking_core using (active);
create policy core_reads_announcement_receipts on public.spot_announcements
  for select to parking_core using (true);
create policy core_inserts_announcement_receipts on public.spot_announcements
  for insert to parking_core with check (true);
create policy core_reads_spot_results on public.spots
  for select to parking_core using (true);
create policy core_announces_spots on public.spots
  for insert to parking_core with check (status = 'announced');
create policy core_records_announced_events on public.outbox
  for insert to parking_core with check (type = 'availability.spot.announced.v1');
