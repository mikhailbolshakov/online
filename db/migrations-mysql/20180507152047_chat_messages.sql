
-- +goose Up
create table if not exists  chat_messages
(
  id                bigint unsigned auto_increment  primary key,
  client_message_id varchar(255)     default ''                               not null,
  chat_id           bigint unsigned                                           not null,
  subscribe_id      bigint unsigned                                           not null,
  type              enum ('message', 'system', 'document') default 'message'  not null,
  message           text             collate 'utf16_unicode_ci'               not null,
  file_id           char(32)         default 0                                not null,
  created_at        timestamp        default CURRENT_TIMESTAMP                not null on update CURRENT_TIMESTAMP,
  updated_at        timestamp        default CURRENT_TIMESTAMP                not null on update CURRENT_TIMESTAMP,
  deleted_at        timestamp                                                 null
);

create index chat_id
  on chat_messages (chat_id);

create index chat_id_client_message_id_index
  on chat_messages (chat_id, client_message_id(32));

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
drop table chat_messages;
