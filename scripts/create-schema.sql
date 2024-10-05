create table if not exists "domains" (
  id integer primary key autoincrement,
  host text unique not null check (host != ''),
  tld text null default null,
  sld text null default null,
  trd text null default null,
  created_at timestamp not null default current_timestamp
);

create table flags (
    domain_id       integer not null,
    flag            text    not null,
    additional_data text,
    created_at timestamp not null default current_timestamp,

    primary key (domain_id, flag),
    foreign key (domain_id) references domains (id)
);

create table crunch_data (
  id integer primary key autoincrement,

  name                        text      not null,
  number_of_employees         integer   not null,
  total_investment_amount_usd integer   not null,
  last_investment_amount_usd  integer   not null,

  -- links
  crunchbase_url text      not null unique,
  linkedin_url   text      not null,
  website_url    text      not null,

  website_host text not null,

  -- dates
  founded_at                  timestamp not null,

  created_at timestamp not null default current_timestamp
);

create table crunch_data_domains (
  crunch_data_id integer not null,
  domain_id      integer not null,

  primary key (crunch_data_id, domain_id),
  foreign key (crunch_data_id) references crunch_data (id),
  foreign key (domain_id)      references domains (id)
);

alter table domains add column tld_sld_trd text generated always as (trd || '.' || sld || '.' || tld);
alter table domains add column tld_sld text generated always as (sld || '.' || tld);

