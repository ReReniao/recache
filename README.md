# 分布式缓存系统
- 并发访问控制（singleFlight）
- 负载均衡（consistenthash 算法）
- 通过回调函数实现无缓存数据请求的响应
- 多种缓存淘汰策略（lru、lfu、fifo，策略类模式）
- 分布式缓存节点间基于http、gRPC协议通信
- TTL 机制，自动清理过期缓存
- 基于etcd实现高可用集群
- 通过endpoint-manager实现服务注册发现
- 缓存穿透防御机制
