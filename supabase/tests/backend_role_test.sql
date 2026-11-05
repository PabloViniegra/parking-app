begin;
set local search_path = public, extensions;
select plan(12);

select has_role('parking_core', 'the API has a dedicated database role');

select ok(
  not rolsuper and not rolcreaterole and not rolcreatedb and not rolreplication and not rolbypassrls and not rolinherit,
  'the backend role cannot bypass RLS or escalate privileges'
) from pg_roles where rolname = 'parking_core';

select ok(
  not has_table_privilege('parking_core', 'public.' || table_name, 'UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER, MAINTAIN'),
  format('the backend role cannot modify or delete existing rows in %s', table_name)
) from (values ('coverage_areas'), ('spots'), ('outbox'), ('timers'), ('spot_daily_counts'), ('spot_announcements')) as tables(table_name);

select ok(not has_column_privilege('parking_core', 'public.spots', 'location', 'SELECT'), 'the backend role cannot read stored precise locations');
select ok(not has_table_privilege('parking_core', 'public.outbox', 'SELECT'), 'the backend role cannot read event history');
select ok(not has_function_privilege('parking_core', 'private.purge_spot_locations(timestamptz)', 'EXECUTE'), 'the backend role cannot run retention maintenance');
select ok(
  not pg_has_role('anon', 'parking_core', 'MEMBER')
  and not pg_has_role('authenticated', 'parking_core', 'MEMBER')
  and not pg_has_role('authenticator', 'parking_core', 'MEMBER'),
  'Data API roles cannot assume the backend role'
);

select * from finish();
rollback;
