CREATE TABLE IF NOT EXISTS `message_history`(
    `id`         BIGINT UNSIGNED AUTO_INCREMENT,
    `message`    text        NOT NULL,
    `tip`        text,
    `exp`        text,
    `resp`       text,
    `home_id`    VARCHAR(40) NOT NULL,
    `updated_at` timestamp   NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据修改时间',
    `created_at` timestamp   NULL DEFAULT CURRENT_TIMESTAMP COMMENT '数据入库时间',
    PRIMARY KEY (`id`),
    INDEX index_created_home (`created_at`, `home_id`)
) ENGINE = InnoDB AUTO_INCREMENT = 100000  COMMENT ='家庭发送消息历史记录表' DEFAULT CHARSET = utf8;
