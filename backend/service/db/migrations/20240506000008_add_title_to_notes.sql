-- 增加标题字段到笔记表
ALTER TABLE notes ADD COLUMN title VARCHAR(255) NOT NULL DEFAULT '' AFTER user_id;
-- 为标题字段增加索引以提升搜索性能
CREATE INDEX idx_notes_title ON notes(title);
