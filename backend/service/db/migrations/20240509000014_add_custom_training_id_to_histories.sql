-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE training_histories
ADD COLUMN custom_training_id INT UNSIGNED NULL COMMENT '自定义训练ID，仅自定义训练类型使用' AFTER training_type,
ADD INDEX idx_custom_training_id (custom_training_id);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back

ALTER TABLE training_histories
DROP INDEX idx_custom_training_id,
DROP COLUMN custom_training_id;
