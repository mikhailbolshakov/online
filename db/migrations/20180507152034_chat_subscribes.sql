
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
create table if not exists chat_subscribes
(
  id         bigint unsigned auto_increment primary key,
  chat_id    bigint unsigned                                        not null,
  active     tinyint unsigned default 1                             not null,
  user_id    bigint unsigned                                        not null,
  user_type  enum ('client', 'operator', 'doctor') default 'client' not null,
  created_at timestamp default CURRENT_TIMESTAMP                    not null on update CURRENT_TIMESTAMP,
  updated_at timestamp default CURRENT_TIMESTAMP                    not null on update CURRENT_TIMESTAMP,
  deleted_at timestamp                                              null
);

create index user_id_active
  on chat_subscribes (user_id, active);

create index chat_id_user_id
  on chat_subscribes (chat_id, user_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
drop table chat_subscribes;
