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
db40.dbx102.0:				# <required> 点名称，ASCII编码，便于跨语言使用, 可以自定义
  address: ""				# <required> point在设备中的物理地址，程序用来访问实时数据
  label: start				# <optional> 短名称
  name: 启动		   		   # <optional> 长名称
  unit: 单位				   # <optional> 单位字符串
  point_type: 				# <optional> 点类型 1：模拟量，2：状态量, 3:整形量，默认为模拟量
  tags: ["test"]			# <optional> tag标记
  option:					# <optional> 状态量映射关系
    1: 启动
    0: 停止
  parameter: 0.707423		# <optional> 模拟量系数
  extra:					# <optional> 扩展字段
  	key1: value1
  	key2: value2
```



# 控制功能

>接口：POST
>
>Path：/point_value/:deviceName
>
>Content-Type：applicant/json
>
>参数：{"TestTag1", 2,  "TestTag2": true, "TestTag3":  1.123}

示例：

```http
POST /point_value/opc HTTP/1.1
Host: localhost:9999
Content-Type: application/json
{
	"Channel1.Device1.Tag1": 0,
	"Tag2": 1.1,
	"Tag3": false
}
```



# 报警功能

> 报警功能走WebSocket接口
>
> Path：/interface/alarm 
>
> 客户端建立连接后等待接收报警信息

报警信息结构：

```json
{
  "name":"TestAlarm",
  "input_name":"inputs.opc",
  "timestamp":"2018-12-15T12:08:03+08:00"
}
```



