--liquibase formatted sql logicalFilePath:changelog-v1.0.0.sql


--changeset hou yong:1006
--comment: add  cut record table
CREATE TABLE  IF NOT EXISTS cut_record (
                                    id varchar(100) NOT NULL,
                                    `type` varchar(20) NULL,
                                    request TEXT NULL,
                                    response TEXT NULL,
                                    create_time DATETIME NULL,
                                    user_id varchar(100) NOT NULL,
                                    code varchar(100) NULL,
                                    name varchar(100) NOT NULL,
                                    CONSTRAINT cut_record_pk PRIMARY KEY (id),
                                    INDEX idx_create_time (create_time),
                                    INDEX idx_type (`type`),
                                    INDEX idx_user_id (user_id)
)
    ENGINE=InnoDB
DEFAULT CHARSET=utf8
COLLATE=utf8_general_ci;

