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
    `device_zn`   VARCHAR(30) COMMENT '设备名称中文',
    `device_en`   VARCHAR(30) COMMENT '设备名称英文',
    `device_id`   VARCHAR(30) COMMENT '设备id',
    `device_des`  VARCHAR(200) COMMENT '设备描述',
    `version`     VARCHAR(30) COMMENT '版本',
    `user_id`     VARCHAR(30) COMMENT '用户id',
    `control`     VARCHAR(30) COMMENT '控制',
    `delay_time`  VARCHAR(30)    DEFAULT "0" COMMENT '延时的时间，单位是秒',
    `ip`          VARCHAR(30) COMMENT 'ip',
    `wifi`        VARCHAR(100) COMMENT 'wifi',
    `updated_at`  timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据修改时间',
    `created_at`  timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '数据入库时间',
    PRIMARY KEY (`id`),
    INDEX index_created_device (`created_at`, `device_id`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 100000 COMMENT ='设备列表'
  DEFAULT CHARSET = utf8;

ALTER TABLE `device`
    ADD COLUMN `delay` TINYINT COMMENT '是否延时，0不延时，1延时' AFTER `control`;
ALTER TABLE `device`
    ADD COLUMN `delay_time` INT COMMENT '延时的时间，单位是秒' AFTER `delay`;


ALTER TABLE `device`
    MODIFY device_type VARCHAR(30);
ALTER TABLE `device`
    MODIFY control VARCHAR(30);
ALTER TABLE `device`
    MODIFY delay_time VARCHAR(200);
ALTER TABLE `device`
    DROP COLUMN delay;

DROP TABLE `user`;
CREATE TABLE IF NOT EXISTS `user`
(
    `id`           BIGINT UNSIGNED AUTO_INCREMENT,
    `phone`        VARCHAR(20),
    `ha_address_1` VARCHAR(200) COMMENT '公网ip加端口',
    `ha_address_2` VARCHAR(200) COMMENT '内网ip加端口',
    `updated_at`   timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '数据修改时间',
    `created_at`   timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '数据入库时间',
    PRIMARY KEY (`id`),
    INDEX index_created_phone (`created_at`, `phone`)
) ENGINE = InnoDB
  AUTO_INCREMENT = 30000 COMMENT ='用户信息表'
  DEFAULT CHARSET = utf8;