CREATE USER test_user WITH PASSWORD 'test_password' CREATEDB;

CREATE SCHEMA IF NOT EXISTS go_template AUTHORIZATION test_user;

ALTER DATABASE test SET SEARCH_PATH TO go_template;

ALTER USER test_user SET SEARCH_PATH TO go_template;

GRANT ALL PRIVILEGES ON DATABASE test TO test_user;
GRANT ALL PRIVILEGES ON SCHEMA go_template TO test_user;

ALTER DEFAULT PRIVILEGES FOR USER test_user IN SCHEMA go_template
    GRANT ALL PRIVILEGES ON TABLES TO test_user;

ALTER DEFAULT PRIVILEGES FOR USER test_user IN SCHEMA go_template
    GRANT ALL PRIVILEGES ON SEQUENCES TO test_user;

-- user and schema for integration testing
CREATE USER test_user_it WITH PASSWORD 'test_password_it' CREATEDB;

CREATE SCHEMA IF NOT EXISTS go_template_it AUTHORIZATION test_user_it;

ALTER USER test_user_it SET SEARCH_PATH TO go_template_it;

GRANT ALL PRIVILEGES ON DATABASE test TO test_user_it;
GRANT ALL PRIVILEGES ON SCHEMA go_template_it TO test_user_it;

ALTER DEFAULT PRIVILEGES FOR USER test_user_it IN SCHEMA go_template_it
    GRANT ALL PRIVILEGES ON TABLES TO test_user_it;

ALTER DEFAULT PRIVILEGES FOR USER test_user_it IN SCHEMA go_template_it
    GRANT ALL PRIVILEGES ON SEQUENCES TO test_user_it;
