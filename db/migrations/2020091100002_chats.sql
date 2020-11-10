-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
create table chat_statuses
(
  code varchar(64) primary key,
  description varchar not null
);
insert into chat_statuses values('opened', 'открыт');
insert into chat_statuses values('closed', 'закрыт');

create table chats
(
  id         uuid primary key,
  reference_id  varchar not null,
  status     varchar references chat_statuses(code) not null,
  created_at timestamp default CURRENT_TIMESTAMP        not null,
  updated_at timestamp default CURRENT_TIMESTAMP        not null,
  deleted_at timestamp                                  null
);

create index idx_chats_reference_id on chats(reference_id);

create table chat_subscribes
(
  id         uuid primary key,
  chat_id    uuid not null,
  account_id uuid not null,
  role       varchar not null,
  created_at timestamp default CURRENT_TIMESTAMP not null,
  updated_at timestamp default CURRENT_TIMESTAMP not null,
  deleted_at timestamp null
);

create index idx_chat_subscr_account_id on chat_subscribes(account_id);
create index idx_chat_subscr_chat_account on chat_subscribes(chat_id, account_id);

create table chat_message_types (
  code varchar(64) primary key,
  description varchar not null
);
insert into chat_message_types values('message', 'текстовое сообщение');
insert into chat_message_types values('file', 'файл');

create table chat_messages
(
  id                uuid primary key,
  client_message_id varchar not null,
  chat_id           uuid not null,
  subscribe_id      uuid not null,
  type              varchar references chat_message_types(code) not null,
  message           text,
  file_id           varchar,
  params 			json,
  read              smallint check(read in (0, 1)) not null,
  created_at        timestamp  default CURRENT_TIMESTAMP not null,
  updated_at        timestamp  default CURRENT_TIMESTAMP not null,
  deleted_at        timestamp null
);
create index idx_chat_msg_chat on chat_messages(chat_id);
create index idx_chat_msg_client_message on chat_messages (client_message_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
drop table chat_messages;
drop table chat_message_statuses;
drop table chat_message_types;
drop table chat_subscribes;
drop table chats;
drop table chat_statuses;