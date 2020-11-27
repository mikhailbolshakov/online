-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
create table rooms
(
  id         uuid primary key,
  reference_id  varchar not null,
  hash varchar not null,
  chat smallint check(chat in (0, 1)) not null,
  video smallint check(video in (0, 1)) not null,
  audio smallint check(audio in (0, 1)) not null,
  closed_at timestamp null,
  created_at timestamp default CURRENT_TIMESTAMP        not null,
  updated_at timestamp default CURRENT_TIMESTAMP        not null,
  deleted_at timestamp                                  null
);

alter table rooms add constraint uk_rooms_ref unique(reference_id, deleted_at);
create index idx_rooms_reference_id on rooms(reference_id);

create table room_subscribers
(
  id         uuid primary key,
  room_id    uuid not null,
  account_id uuid not null,
  role       varchar not null,
  system_account smallint check(system_account in (0, 1)) not null,
  unsubscribe_at timestamp null,
  created_at timestamp default CURRENT_TIMESTAMP not null,
  updated_at timestamp default CURRENT_TIMESTAMP not null,
  deleted_at timestamp null
);

create index idx_room_subscr_account_id on room_subscribers(account_id);
alter table room_subscribers add constraint uk_room_subscr_room_acc unique (room_id, account_id);

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
  room_id           uuid not null,
  subscribe_id      uuid not null,
  account_id        uuid not null,
  type              varchar references chat_message_types(code) not null,
  message           text,
  file_id           varchar,
  params 			json,
  recipient_account_id uuid null,
  created_at        timestamp  default CURRENT_TIMESTAMP not null,
  updated_at        timestamp  default CURRENT_TIMESTAMP not null,
  deleted_at        timestamp null
);
create index idx_chat_msg_chat on chat_messages(room_id);
create index idx_chat_msg_client_message on chat_messages (client_message_id);
create index idx_chat_msg_subscr on chat_messages(subscribe_id);
create index idx_chat_msg_acc on chat_messages(account_id);
create index idx_chat_msg_rec_acc on chat_messages(recipient_account_id);

create table chat_message_statuses
(
  id            uuid primary key,
  message_id    uuid not null,
  subscribe_id  uuid not null,
  account_id    uuid not null,
  status        varchar check (status in ('recd', 'read')) default 'recd' not null,
  created_at    timestamp default CURRENT_TIMESTAMP          not null,
  updated_at    timestamp default CURRENT_TIMESTAMP          not null,
  deleted_at    timestamp                                    null
);

create index idx_chat_mes_statuses_message_id on chat_message_statuses(message_id);
create index idx_chat_mes_statuses_subscribe_id on chat_message_statuses(subscribe_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
drop table chat_message_statuses;
drop table chat_messages;
drop table chat_message_types;
drop table room_subscribers;
drop table rooms;
