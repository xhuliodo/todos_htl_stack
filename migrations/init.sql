CREATE TABLE IF NOT EXISTS todos(
  id integer PRIMARY KEY AUTOINCREMENT,
  content text NOT NULL,
  created_at datetime DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime DEFAULT CURRENT_TIMESTAMP,
  completed boolean DEFAULT FALSE,
  user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS users(
  id integer PRIMARY KEY AUTOINCREMENT,
  name text NOT NULL,
  email text NOT NULL UNIQUE,
  password text NOT NULL,
  created_at datetime DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime DEFAULT CURRENT_TIMESTAMP
);

