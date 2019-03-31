CREATE TABLE `owners`
(
  `owner_id`   varchar(33) NOT NULL,
  `created_at` bigint(20) unsigned NOT NULL,
  `updated_at` bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `events`
(
  `id`         int(20) NOT NULL AUTO_INCREMENT,
  `event_id`   varchar(30) NOT NULL,
  `created_at` bigint(20) unsigned NOT NULL,
  `updated_at` bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_event_id` (`event_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `event_statuses`
(
  `owner_id`   varchar(33) NOT NULL,
  `event_id`   varchar(30) NOT NULL,
  `status`     int(1) NOT NULL,
  `created_at` bigint(20) unsigned NOT NULL,
  `updated_at` bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`owner_id`, `event_id`),
  KEY `idx_owner_id` (`owner_id`),
  KEY `idx_event_id` (`event_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `event_participants`
(
  `user_id`         varchar(33) NOT NULL,
  `event_id`        varchar(30) NOT NULL,
  `is_participated` tinyint(1) NOT NULL,
  `created_at`      bigint(20) unsigned NOT NULL,
  `updated_at`      bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`user_id`, `event_id`),
  KEY `user_id` (`user_id`),
  KEY `idx_is_participated` (`is_participated`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `event_votes`
(
  `event_id`   varchar(30) NOT NULL,
  `user_id`    varchar(33) NOT NULL,
  `vote`       int(1) NOT NULL,
  `created_at` bigint(20) unsigned NOT NULL,
  `updated_at` bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`event_id`, `user_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
