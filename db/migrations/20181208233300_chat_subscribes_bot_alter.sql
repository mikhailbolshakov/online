
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
alter table chat_subscribes modify user_type enum('client', 'operator', 'doctor', 'bot') default 'client' not null;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
alter table chat_subscribes modify user_type enum('client', 'operator', 'doctor') default 'client' not null;
