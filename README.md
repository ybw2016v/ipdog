# ipdog

一个可以返回IP地址及其属地的小工具，主要目的是学习GO语言。

利用github上[开源的ip库](https://github.com/a76yyyy/ipdata/releases/tag/v2022.04.29)，使用`gin`作为web框架，`go-sqlite3`读取sqlite数据库，使用`go-redis`连接redis，作为缓存与限流。支持GET和POST两种请求方式，支持日最大访问次数限制，超过限制后随机返回虚假的IP属地。

