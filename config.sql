CREATE TABLE `cloud_configs` (
  `id` int NOT NULL AUTO_INCREMENT,
  `config_key` varchar(255) NOT NULL,
  `config_value` longtext NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `delete_at` timestamp NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_config_key` (`config_key`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='cloud config table'