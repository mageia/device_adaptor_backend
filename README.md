# 统一数据监控平台

主要为解决以下问题而设计：

- 避免从OPC取数据，直接从PLC获取数据
- 统一控制流程问题，指令的发送、调度、回调等
- 统一数据格式问题，避免维护多种输出或输出的类型、格式等



# 数据输出格式

每个数据源的每个采集周期输出数据格式如下：

- fields:：数据字段
- name：数据源标识
- quality：数据质量
  - 1：Good
  - 2：Disconnect
  - 。。。：待扩展
  - MaxUint8：Unknown
- timestamp：ms时间戳

```json
[{
"fields":{
    "db40.dbd2":2.87,
    "db40.dbd6":3.46,
    "db40.dbw0":123,
    "db40.dbw112":1234,
    "db40.dbx102.0":1,
    "db40.dbx102.1":1,
    "db40.dbx102.7":0,
    "db40.dbx104.1":1
},
"name":"s7",
"quality":1,
"timestamp":1540261475146
}]
```



# 点表格式

```yaml
db40.dbx102.0:				# <required> point在设备中的物理地址，唯一标识
  label: start				# <optional> 短名称
  name: 启动		   		   # <optional> 长名称
  unit: 单位				   # <optional> 单位字符串
  point_type: 				# <optional> 点类型 0：模拟量，1：状态量, 默认为模拟量
  tags: ["test"]			# <optional> tag标记
  option:					# <optional> 状态量映射关系
    1: 启动
    0: 停止
  parameter: 0.707423		# <optional> 模拟量系数
  extra:					# <optional> 扩展字段
  	key1: value1
  	key2: value2
```



# Plugins

- Controller
  - [x] http
  - [ ] redis
  - [ ] websocket
- Input
  - [x] ftp
  - [x] modbus
  - [x] s7
- Output
  - [x] file
  - [x] redis
- Parser
  - [x] csv
- Serializer
  - [x] json



## TODO:

- [x] 配置流程Server化
- [x] 配置更新后，程序Reload功能
- [x] 输出数据格式
- [ ] 前端原型
- [ ] 配置管理页面开发
- [ ] 网关自检逻辑开发
- [ ] Metric统计逻辑开发
- [ ] 数据基础分析逻辑开发