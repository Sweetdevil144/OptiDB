DROP ROLE IF EXISTS profiler_ro;
CREATE ROLE profiler_ro WITH LOGIN PASSWORD 'profiler_ro_pass';

GRANT CONNECT ON DATABASE optidb TO profiler_ro;
GRANT USAGE ON SCHEMA public TO profiler_ro;
GRANT SELECT ON pg_stat_statements TO profiler_ro;
GRANT SELECT ON pg_stat_user_tables TO profiler_ro;
GRANT SELECT ON pg_stat_user_indexes TO profiler_ro;
GRANT SELECT ON pg_class TO profiler_ro;
GRANT SELECT ON pg_index TO profiler_ro;

DROP ROLE IF EXISTS profiler_sb;
CREATE ROLE profiler_sb WITH LOGIN PASSWORD 'profiler_sb_pass';

GRANT CONNECT ON DATABASE optidb TO profiler_sb;
GRANT USAGE ON SCHEMA public TO profiler_sb;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO profiler_sb;
GRANT TEMPORARY ON DATABASE optidb TO profiler_sb;
