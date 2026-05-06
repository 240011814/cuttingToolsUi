-- +goose Up
-- +goose StatementBegin
INSERT INTO `permissions` (`code`, `name`, `group_name`) VALUES
('ai:emergency:view', '应急训练', '练习管理')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `group_name` = VALUES(`group_name`);

-- 为 R_SUPER 自动分配该权限
INSERT INTO `role_permissions` (`role_code`, `permission_id`)
SELECT 'R_SUPER', id FROM `permissions` WHERE `code` = 'ai:emergency:view'
ON DUPLICATE KEY UPDATE `role_code` = `role_code`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM `permissions` WHERE `code` = 'ai:emergency:view';
-- +goose StatementEnd
