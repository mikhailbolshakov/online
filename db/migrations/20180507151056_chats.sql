
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
create table if not exists chats
(
  id         bigint unsigned auto_increment             primary key,
  order_id   bigint unsigned                            not null,
  status     enum ('opened', 'closed') default 'opened' not null,
  created_at timestamp default CURRENT_TIMESTAMP        not null on update CURRENT_TIMESTAMP,
  updated_at timestamp default CURRENT_TIMESTAMP        not null on update CURRENT_TIMESTAMP,
  deleted_at timestamp                                  null
);

create index order_id
  on chats (order_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
drop table chats;
