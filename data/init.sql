-- SQLite数据库初始化脚本
-- 创建用户表
CREATE TABLE IF NOT EXISTS `user` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `name` TEXT NOT NULL,
  `status` INTEGER NOT NULL DEFAULT 0,
  `age` INTEGER NOT NULL DEFAULT 0,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建MQTT消息表（如果需要持久化MQTT消息）
CREATE TABLE IF NOT EXISTS `mqtt_message` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `topic` TEXT NOT NULL,
  `payload` TEXT NOT NULL,
  `qos` INTEGER NOT NULL DEFAULT 0,
  `retained` INTEGER NOT NULL DEFAULT 0,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 插入一些测试数据
INSERT OR IGNORE INTO `user` (`id`, `name`, `status`, `age`) VALUES
(1, 'Alice', 0, 25),
(2, 'Bob', 1, 30),
(3, 'Charlie', 0, 28);
