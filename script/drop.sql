-- 创建数据库 kv，如果不存在
CREATE DATABASE IF NOT EXISTS cache_slow_db;

-- 切换到数据库 cache_slow_db
USE cache_slow_db;

-- 删除已存在的 students 表，如果存在
DROP TABLE IF EXISTS `students`;