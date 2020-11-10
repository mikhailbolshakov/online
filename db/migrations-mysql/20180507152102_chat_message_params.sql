
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
create table if not exists chat_message_params
(
  id            bigint unsigned auto_increment primary key,
  message_id    bigint unsigned                              not null,
  `key`         varchar(255)                                 not null,
  value         text                                         not null,
  created_at    timestamp default CURRENT_TIMESTAMP          not null on update CURRENT_TIMESTAMP,
  updated_at    timestamp default CURRENT_TIMESTAMP          not null on update CURRENT_TIMESTAMP,
  deleted_at    timestamp                                    null
);

create index message_id
  on chat_message_params (message_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
drop table chat_message_params;
