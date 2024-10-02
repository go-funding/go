create table if not exists "domains" (
  id integer primary key autoincrement,
  host text unique not null check (host != ''),
  created_at timestamp not null default current_timestamp
);
