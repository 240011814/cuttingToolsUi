-- +goose Up
-- +goose StatementBegin
CREATE TABLE reminders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    remind_at DATETIME NOT NULL,
    notified TINYINT(1) DEFAULT 0,
    repeat_type VARCHAR(20) DEFAULT 'none' COMMENT 'none/daily/weekly/monthly/yearly',
    repeat_interval INT DEFAULT 1,
    repeat_end_at DATETIME NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_remind_at (remind_at),
    INDEX idx_notified (notified)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='备忘提醒';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reminders;
-- +goose StatementEnd
