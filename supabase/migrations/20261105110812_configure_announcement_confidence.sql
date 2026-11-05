update public.coverage_areas
set policies = policies || '{"confirmed_departure_base":0.8,"new_releaser_reliability":0.8}'::jsonb
where slug = 'madrid-region';
