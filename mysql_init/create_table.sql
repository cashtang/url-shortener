CREATE DATABASE IF NOT EXISTS shortener;
use shortener;
CREATE TABLE `shortened_urls` (
  `id` VARCHAR(16) NOT NULL,
  `appid` VARCHAR(32) NOT NULL,
  `long_url` VARCHAR(255) NOT NULL,
  `created` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  `expire` BIGINT NOT NULL,
  PRIMARY KEY  (`id`),
  UNIQUE KEY `long` (`long_url`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE INDEX `shortened_urls_appid` ON `shortened_urls`.`appid`;

CREATE TABLE `auth_apps` (
    `appid` VARCHAR(32) NOT NULL,
    `secret` VARCHAR(64) NOT NULL,
    `disabled` TINYINT DEFAULT 0,
    PRIMARY KEY (`appid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;