### connection-pool
### 介绍
`connection-pool`是基于golang实现的连接池，让调用者在使用中间件的连接时，达到限制过多连接的问题。

![](https://github.com/studyplace-io/connection-pool/tree/main/image?raw=true)

### 项目功能
- 自定义连接数量
- 自定义获取连接超时时间
- 自定义空闲连接时间(超过时间会内部自动回收连接)
- 自定义心跳检查时间(内部定时检查心跳与检查连接数量)
- 支持mysql redis连接池

### 使用
