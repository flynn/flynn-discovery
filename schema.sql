CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE clusters (
  cluster_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  creator_ip text NOT NULL,
  creator_user_agent text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE instances (
  instance_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  cluster_id uuid NOT NULL REFERENCES clusters (cluster_id),
  flynn_version text NOT NULL,
  ssh_public_keys json NOT NULL,
  url text NOT NULL,
  name text NOT NULL,
  creator_ip text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(cluster_id, url)
);
