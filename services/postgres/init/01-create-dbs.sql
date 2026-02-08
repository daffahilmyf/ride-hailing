DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'users') THEN
    CREATE DATABASE users;
  END IF;
END
$$;
