revoke all on all tables in schema public from public, anon, authenticated;

alter default privileges for role postgres in schema public
  revoke all on tables from public, anon, authenticated;

alter default privileges for role postgres in schema public
  revoke all on sequences from public, anon, authenticated;

alter default privileges for role postgres in schema public
  revoke all on functions from anon, authenticated;

-- A schema-level revoke cannot remove PostgreSQL's global PUBLIC execute default.
alter default privileges for role postgres
  revoke execute on functions from public;
