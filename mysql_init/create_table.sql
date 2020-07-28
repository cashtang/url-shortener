CREATE DATABASE IF NOT EXISTS shortener;
use shortener;
CREATE TABLE `shortened_urls` (
  `id` VARCHAR(16) NOT NULL,
  `long_url` varchar(255) NOT NULL,
  `created` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  PRIMARY KEY  (`id`),
  UNIQUE KEY `long` (`long_url`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;