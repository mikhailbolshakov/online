-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
create table account_statuses
(
  code varchar(64) primary key,
  description varchar not null
);
insert into account_statuses values('active', 'активный');
insert into account_statuses values('locked', 'заблокирован');

create table account_types
(
  code varchar(64) primary key,
  description varchar not null
);
insert into account_types values('bot', 'бот');
insert into account_types values('anonymous_user', 'анонимный пользователь');
insert into account_types values('user', 'пользователь');

create table accounts
(
  id uuid primary key,
  account_type varchar references account_types(code) not null,
  status varchar references account_statuses(code) not null,
  account varchar not null,
  external_id varchar null,
  first_name varchar,
  middle_name varchar,
  last_name varchar,
  email varchar,
  phone varchar,
  avatar_url varchar,
  created_at timestamp default CURRENT_TIMESTAMP        not null,
  updated_at timestamp default CURRENT_TIMESTAMP        not null,
  deleted_at timestamp                                  null
);

create index idx_account_account on accounts(account);
create index idx_account_ext_id on accounts(external_id);
create index idx_account_phone on accounts(phone);

create table online_status_types
(
  code varchar(64) primary key,
  description varchar not null
);
insert into online_status_types values('offline', 'оффлайн');
insert into online_status_types values('online', 'онлайн');
insert into online_status_types values('busy', 'занят');
insert into online_status_types values('away', 'отошел');

create table online_statuses
(
  id uuid primary key,
  account_id uuid not null,
  status varchar references online_status_types(code) not null,
  created_at timestamp default CURRENT_TIMESTAMP        not null,
  updated_at timestamp default CURRENT_TIMESTAMP        not null,
  deleted_at timestamp                                  null
);

create index idx_online_status_account_id on online_statuses(account_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
drop table online_statuses;
drop table online_status_types;
drop table accounts;
drop table account_types;
drop table account_statuses;
