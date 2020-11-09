
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE `chat_messages`
CHANGE `type` `type` enum('message','system','document','doctor_detail') COLLATE 'utf8mb4_general_ci' NOT NULL DEFAULT 'message' AFTER `subscribe_id`;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE `chat_messages`
CHANGE `type` `type` enum('message','system','document') COLLATE 'utf8mb4_general_ci' NOT NULL DEFAULT 'message' AFTER `subscribe_id`;
