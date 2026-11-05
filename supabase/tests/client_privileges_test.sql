begin;

create extension if not exists pgtap with schema extensions;
set local search_path = public, extensions;

select plan(22);

select is(current_user::text, 'postgres', 'new objects use the postgres default privileges');

select ok(
  not has_table_privilege(role_name, 'public.' || table_name, 'SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER, MAINTAIN'),
  format('%s has no direct privileges on %s', role_name, table_name)
)
from (values ('anon'), ('authenticated')) as roles(role_name)
cross join (values ('coverage_areas'), ('spots'), ('outbox'), ('timers'), ('spot_daily_counts'), ('spot_announcements')) as tables(table_name);

create table public.client_privileges_probe (id bigint);
alter table public.client_privileges_probe enable row level security;

select ok(
  not has_table_privilege(role_name, 'public.client_privileges_probe', 'SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER, MAINTAIN'),
  format('%s receives no privileges on new tables created by postgres', role_name)
)
from (values ('anon'), ('authenticated')) as roles(role_name);

create sequence public.client_privileges_probe_sequence;

select ok(
  not has_sequence_privilege(role_name, 'public.client_privileges_probe_sequence', 'USAGE, SELECT, UPDATE'),
  format('%s receives no privileges on new sequences created by postgres', role_name)
)
from (values ('anon'), ('authenticated')) as roles(role_name);

create function public.client_privileges_probe_function()
returns integer language sql security invoker set search_path = ''
as $$ select 1 $$;

select ok(
  not has_function_privilege(role_name, 'public.client_privileges_probe_function()', 'EXECUTE'),
  format('%s receives no execute privilege on new functions created by postgres', role_name)
)
from (values ('anon'), ('authenticated')) as roles(role_name);

select ok(
  has_table_privilege('service_role', 'public.client_privileges_probe', 'SELECT')
  and has_table_privilege('service_role', 'public.client_privileges_probe', 'INSERT'),
  'service_role retains access to new tables'
);

select ok(
  has_sequence_privilege('service_role', 'public.client_privileges_probe_sequence', 'USAGE'),
  'service_role retains access to new sequences'
);

select ok(
  has_function_privilege('service_role', 'public.client_privileges_probe_function()', 'EXECUTE'),
  'service_role retains access to new public functions'
);

select * from finish();
rollback;
