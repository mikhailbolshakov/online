
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE `chat_messages`
CHANGE `type` `type` enum('message','system','document','doctor_detail', 'clinic_detail', 'pay_detail', 'product_detail', 'medcard_detail', 'order_detail', 'promocode_link', 'product_link')
COLLATE 'utf8mb4_general_ci' NOT NULL DEFAULT 'message' AFTER `subscribe_id`;

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE `chat_messages`
CHANGE `type` `type` enum('message','system','document','doctor_detail', 'clinic_detail', 'pay_detail', 'product_detail', 'medcard_detail', 'order_detail')
COLLATE 'utf8mb4_general_ci' NOT NULL DEFAULT 'message' AFTER `subscribe_id`;
