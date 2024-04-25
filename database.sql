DROP TABLE `message_history`;
CREATE TABLE IF NOT EXISTS `message_history`
(
    `id`          BIGINT UNSIGNED AUTO_INCREMENT,
    `option`      tinyint,
    `merchant_id` VARCHAR(30),
    `message`     text      NOT NULL,
    `tip`         text,
    `exp`         text,
    `resp`        text,
    `identity`    VARCHAR(40) COMMENT '用户身份标识',
    `updated_at`  timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据修改时间',
    `created_at`  timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '数据入库时间',
    PRIMARY KEY (`id`),
    INDEX index_created_home (`created_at`, `identity`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 100000 COMMENT ='家庭发送消息历史记录表'
  DEFAULT CHARSET = utf8;

DROP TABLE `device`;
CREATE TABLE IF NOT EXISTS `device`
(
    `id`             BIGINT UNSIGNED AUTO_INCREMENT,
    `device_id`      VARCHAR(30) NOT NULL,
    `merchant_id`    VARCHAR(30) NOT NULL,
    `home_id`        VARCHAR(30) NOT NULL,
    `device_name`    VARCHAR(30) NOT NULL,
    `device_chinese` VARCHAR(30) NOT NULL,
    `capability`     VARCHAR(30) NOT NULL,
    `command`        VARCHAR(30) NOT NULL,
    `arguments`      VARCHAR(30) NOT NULL,
    `ranges`         VARCHAR(30) NOT NULL,
    `value`          VARCHAR(30) NOT NULL,
    `updated_at`     timestamp   NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据修改时间',
    `created_at`     timestamp   NULL DEFAULT CURRENT_TIMESTAMP COMMENT '数据入库时间',
    PRIMARY KEY (`id`),
    INDEX index_created_home (`home_id`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 100000 COMMENT ='家庭设备信息记录表'
  DEFAULT CHARSET = utf8;

DROP TABLE `device`;
CREATE TABLE IF NOT EXISTS `device`
(
    `id`          BIGINT UNSIGNED AUTO_INCREMENT,
    `device_type` VARCHAR(30) COMMENT '设备类型',
    `device_name` VARCHAR(30) NOT NULL COMMENT '设备名称',
    `device_id`   VARCHAR(30) NOT NULL COMMENT '设备id',
    `user_id`     VARCHAR(30) NOT NULL COMMENT '用户id',
    `action`      VARCHAR(30) COMMENT '设备动作',
    `timestamp`   BIGINT COMMENT '设备记录时间戳',
    `data`        text COMMENT '设备自带上传的数据',
    `updated_at`  timestamp   NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据修改时间',
    `created_at`  timestamp   NULL DEFAULT CURRENT_TIMESTAMP COMMENT '数据入库时间',
    PRIMARY KEY (`id`),
    INDEX index_created_device (`created_at`, `device_id`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 100000 COMMENT ='设备列表'
  DEFAULT CHARSET = utf8;