BEGIN transaction;

create extension IF NOT EXISTS  citext;
create extension IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS pills;


CREATE TABLE IF NOT EXISTS pills.users (
  user_id bigint primary key not null,
  username text,
  first_name text,
  last_name text,
  created timestamp with time zone default (now() at time zone 'utc')
);

CREATE TABLE IF NOT EXISTS pills.messages (
  id bigserial primary key not null,
  user_id bigint not null,
  msg_text text not null,
  is_callback_query boolean not null,
  time_sent timestamp with time zone default (now() at time zone 'utc'),
  FOREIGN KEY (user_id) REFERENCES pills.users(user_id)
);

CREATE TABLE IF NOT EXISTS pills.pills (
  pill_id bigserial primary key not null,
  user_id bigint not null,
  pill_name text not null,
  pill_hour int not null,
  pill_min int not null,
  next_remind_time timestamp with time zone  not null,
  FOREIGN KEY (user_id) REFERENCES pills.users(user_id)
);

-- CREATE TABLE IF NOT EXISTS pills.today_reminds (
--   today date not null,
--   user_id bigint not null,
--   pill_name text not null,
--   pill_hour_greenwich int not null,
--   pill_min_greenwich int not null,
--   reminded bool not null default false,
--   FOREIGN KEY (user_id) REFERENCES pills.users(user_id)
-- );


CREATE INDEX IF NOT EXISTS IX_pills_pills ON pills.pills (user_id);

END TRANSACTION;