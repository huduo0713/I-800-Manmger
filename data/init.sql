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

-- 算法表 - 与边缘设备算法表同步
CREATE TABLE IF NOT EXISTS `algorithm` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `algorithm_id` TEXT NOT NULL UNIQUE,
  `algorithm_name` TEXT NOT NULL,
  `algorithm_version` TEXT NOT NULL,
  `algorithm_version_id` TEXT NOT NULL,
  `algorithm_data_url` TEXT NOT NULL,
  `file_size` INTEGER NOT NULL DEFAULT 0,
  `md5` TEXT NOT NULL CHECK(length(md5) = 32),  -- MD5必须是32位
  `local_path` TEXT, -- 本地存储路径
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  
  -- 复合唯一约束：同一算法ID+版本只能有一条记录
  UNIQUE(algorithm_id, algorithm_version)
);

-- 创建触发器：自动更新updated_at时间戳
CREATE TRIGGER IF NOT EXISTS algorithm_updated_at 
AFTER UPDATE ON algorithm 
FOR EACH ROW
BEGIN
  UPDATE algorithm SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- 插入示例算法数据
INSERT OR IGNORE INTO `algorithm` (
  `algorithm_id`, 
  `algorithm_name`, 
  `algorithm_version`, 
  `algorithm_version_id`, 
  `algorithm_data_url`, 
  `file_size`, 
  `md5`
) VALUES (
  'uuid-1234',
  '夜间节能策略算法',
  '1.0',
  'uuid-1234',
  'http://113.249.91.53:9001/haikang/algorithmZip/5fe37a4d248b413d8e62057bc6adb11c',
  5242880,
  'a1b2c3d4e5f67890abcdef1234567890'
);