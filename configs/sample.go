package configs

import "time"

type ConfigSample struct {
	Key     string
	Label   string
	Default interface{}
	Choice  interface{}
	Type    string
	Order   int
}
type ConfigSampleArray []ConfigSample

func (c ConfigSampleArray) Len() int {
	return len(c)
}
func (c ConfigSampleArray) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c ConfigSampleArray) Less(i, j int) bool {
	return c[i].Order < c[j].Order
}

var ParserSample = map[string]map[string]ConfigSample{
	"csv": {
		"header_row_count": ConfigSample{"header_row_count", "文件头行数", 1, nil, "input", 0},
	},
}
var InputSample = map[string]map[string]ConfigSample{
	"_base": {
		"created_at":   ConfigSample{"created_at", "创建时间", time.Now().UnixNano() / 1e6, nil, "none", -100},
		"field_prefix": ConfigSample{"field_prefix", "测点前缀", "", nil, "input", 21},
		"field_suffix": ConfigSample{"field_suffix", "测点后缀", "", nil, "input", 22},
	},
	"serial": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "serial", nil, "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "serial", nil, "input", 2},
		"address":       ConfigSample{"address", "串口地址", "", nil, "input", 3},
		"baud_rate":     ConfigSample{"baud_rate", "波特率", 115200, nil, "input", 4},
		"interactive":   ConfigSample{"interactive", "交互式", true, nil, "radio", 5},
		"parser":        ConfigSample{"parser", "解析器", map[string]interface{}{"vibration": map[string]interface{}{}}, nil, "text", 6},
		"interval":      ConfigSample{"interval", "采集周期", "5s", nil, "combine", 20},
	},
	"eip": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "eip", nil, "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "eip", nil, "input", 2},
		"address":       ConfigSample{"address", "数据源地址", "10.211.55.4:44818", nil, "input", 3},
		"interval":      ConfigSample{"interval", "采集周期", "5s", nil, "combine", 20},
	},
	"opc_tcp": {
		"plugin_name":     ConfigSample{"plugin_name", "插件名称", "opc_tcp", nil, "select", 1},
		"name_override":   ConfigSample{"name_override", "数据源名称", "opc_tcp", nil, "input", 2},
		"address":         ConfigSample{"address", "数据源地址", "10.211.55.18:8090", nil, "input", 3},
		//"opc_server_name": ConfigSample{"opc_server_name", "OPC名称", "Matrikon.OPC.Simulation.1", nil, "input", 4},
		"interval":        ConfigSample{"interval", "采集周期", "5s", nil, "combine", 20},
	},
	"modbus": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "modbus", nil, "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "modbus", nil, "input", 2},
		"address":       ConfigSample{"address", "数据源地址", "10.211.55.4:502", nil, "input", 3},
		"slave_id":      ConfigSample{"slave_id", "从站地址", 1, nil, "input", 4},
		"interval":      ConfigSample{"interval", "采集周期", "3s", nil, "combine", 20},
	},
	"fake": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "fake", nil, "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "fake", nil, "input", 2},
		"interval":      ConfigSample{"interval", "采集周期", "3s", nil, "combine", 20},
	},
	"ftp": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "ftp", nil, "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "ftp", nil, "input", 2},
		"address":       ConfigSample{"address", "地址", "ftp://leaniot:leaniot@localhost:21/SubFile", nil, "input", 5},
		"point_path":    ConfigSample{"point_path", "点表文件路径", "", nil, "input", 6},
		"point_decode":  ConfigSample{"point_decode", "点表文件编码", "utf-8", []string{"gbk", "utf-8"}, "input", 7},
		"data_path":     ConfigSample{"data_path", "数据文件路径", "", nil, "input", 8},
		"data_decode":   ConfigSample{"data_decode", "数据文件编码", "utf-8", []string{"gbk", "utf-8"}, "input", 9},
		"point_parser":  ConfigSample{"point_parser", "点表解析器", nil, nil, "text", 10},
		"data_parser":   ConfigSample{"data_parser", "数据解析器", nil, nil, "text", 11},
		"interval":      ConfigSample{"interval", "采集周期", "3s", nil, "combine", 20},
	},
	"s7": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "s7", nil, "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "s7", nil, "input", 2},
		"address":       ConfigSample{"address", "数据源地址", "192.168.0.168", nil, "input", 3},
		"rack":          ConfigSample{"rack", "机架号", 0, nil, "input", 4},
		"slot":          ConfigSample{"slot", "槽号", 1, nil, "input", 5},
		"interval":      ConfigSample{"interval", "采集周期", "3s", nil, "combine", 20},
	},
	"kj66": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "kj66", nil, "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "kj66", nil, "input", 2},
		"address":       ConfigSample{"address", "数据源地址", "192.168.0.168", nil, "input", 3},
		"version":       ConfigSample{"version", "分站版本", "5", map[string]interface{}{"5": nil, "6": nil, "7": nil}, "select", 3},
		"interval":      ConfigSample{"interval", "采集周期", "3s", nil, "combine", 20},
	},
	"http_listener": {
		"plugin_name":    ConfigSample{"plugin_name", "插件名称", "http_listener", nil, "select", 1},
		"listen_address": ConfigSample{"listen_address", "监听地址", "0.0.0.0:19999", nil, "input", 2},
		"max_body_size":  ConfigSample{"max_body_size", "最大消息体大小", 5 * 1024 * 1024, nil, "input", 3},
		"max_line_size":  ConfigSample{"max_line_size", "最大文件行数", 64 * 1024, nil, "input", 4},
		"read_timeout":   ConfigSample{"read_timeout", "读超时时间", "10s", nil, "combine", 5},
		"write_timeout":  ConfigSample{"write_timeout", "写超时时间", "10s", nil, "combine", 6},
		"basic_username": ConfigSample{"basic_username", "认证账户", "", nil, "input", 7},
		"basic_password": ConfigSample{"basic_password", "认证密码", "", nil, "input", 8},
	},
	"snmp": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "snmp", nil, "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "snmp", nil, "input", 2},
		"address":       ConfigSample{"address", "数据源地址", "192.168.123.253:161", nil, "input", 3},
		"version":       ConfigSample{"version", "版本", "v2c", map[string]interface{}{"v1": 0, "v2c": 1, "v3": 2}, "select", 3},
		"interval":      ConfigSample{"interval", "采集周期", "3s", nil, "combine", 20},
	},
	"ws": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "ws", nil, "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "ws", nil, "input", 2},
		"address":       ConfigSample{"address", "数据源地址", "ws://localhost:8888", nil, "input", 3},
		"parser":        ConfigSample{"parser", "解析器", map[string]interface{}{"kj66": map[string]interface{}{}}, nil, "text", 6},
	},
}
var OutputSample = map[string]map[string]ConfigSample{
	"_base": {
		"metric_buffer_limit": ConfigSample{"metric_buffer_limit", "批量上传缓冲区大小", 0, nil, "input", 100},
		"metric_batch_size":   ConfigSample{"metric_batch_size", "测点批量上传数量", 0, nil, "input", 101},
		"created_at":          ConfigSample{"created_at", "创建时间", time.Now().UnixNano() / 1e6, nil, "none", 0},
	},
	"file": {
		"plugin_name": ConfigSample{"plugin_name", "插件名称", "file", nil, "select", 1},
		"files":       ConfigSample{"files", "输出地址", []string{"stdout"}, nil, "multi-input", 2},
	},
	"redis": {
		"plugin_name":        ConfigSample{"plugin_name", "插件名称", "redis", nil, "select", 1},
		"url_address":        ConfigSample{"url_address", "地址URL", "redis://localhost:6379/0", nil, "input", 2},
		"points_key":         ConfigSample{"points_key", "点表内容 Key", "points", nil, "input", 3},
		"points_version_key": ConfigSample{"points_version_key", "点表版本 Key", "points:version", nil, "input", 4},
	},
	"amqp": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "amqp", nil, "select", 1},
		"url_address":   ConfigSample{"url_address", "地址URL", "amqp://guest:guest@localhost:5672/", nil, "input", 2},
		"exchange_name": ConfigSample{"exchange_name", "交换器名称", "red_gateway", nil, "input", 5},
	},
	"mqtt": {
		"plugin_name": ConfigSample{"plugin_name", "插件名称", "mqtt", nil, "select", 1},
		"url_address": ConfigSample{"url_address", "地址URL", "mqtt://guest:guest@localhost:1883/", nil, "input", 2},
	},
}
var ControllerSample = map[string]map[string]ConfigSample{
	"_base": {
		"created_at": ConfigSample{"created_at", "创建时间", time.Now().UnixNano() / 1e6, nil, "none", 0},
	},
	"http": {
		"plugin_name": ConfigSample{"plugin_name", "插件名称", "http", nil, "select", 1},
		"address":     ConfigSample{"address", "监听地址", "0.0.0.0:9999", nil, "input", 2},
	},
}
