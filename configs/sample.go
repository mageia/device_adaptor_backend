package configs

import "time"

type ConfigSample struct {
	Key     string
	Label   string
	Default interface{}
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

var InputSample = map[string]map[string]ConfigSample{
	"_base": {
		"created_at": ConfigSample{"created_at", "创建时间", time.Now().UnixNano() / 1e6, "none", -100},
	},
	"serial": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "serial", "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "serial", "input", 2},
		"address":       ConfigSample{"address", "串口地址", "", "input", 3},
		"baud_rate":     ConfigSample{"baud_rate", "波特率", 115200, "input", 4},
		"interactive":   ConfigSample{"interactive", "交互式", true, "radio", 5},
		"parser":        ConfigSample{"parser", "解析器", map[string]interface{}{"vibration": map[string]interface{}{}}, "text", 6},
		"interval":      ConfigSample{"interval", "采集周期", "5s", "combine", 20},
		"field_prefix":  ConfigSample{"field_prefix", "测点前缀", "", "input", 21},
		"field_suffix":  ConfigSample{"field_suffix", "测点后缀", "", "input", 22},
	},
	"eip": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "eip", "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "eip", "input", 2},
		"address":       ConfigSample{"address", "数据源地址", "10.211.55.4:44818", "input", 3},
		"interval":      ConfigSample{"interval", "采集周期", "5s", "combine", 20},
		"field_prefix":  ConfigSample{"field_prefix", "测点前缀", "", "input", 21},
		"field_suffix":  ConfigSample{"field_suffix", "测点后缀", "", "input", 22},
	},
	"opc": {
		"plugin_name":     ConfigSample{"plugin_name", "插件名称", "opc", "select", 1},
		"name_override":   ConfigSample{"name_override", "数据源名称", "opc", "input", 2},
		"address":         ConfigSample{"address", "数据源地址", "10.211.55.4:2048", "input", 3},
		"opc_server_name": ConfigSample{"opc_server_name", "OPC名称", "Kepware.KepServerEx.V5", "input", 4},
		"interval":        ConfigSample{"interval", "采集周期", "5s", "combine", 20},
		"field_prefix":    ConfigSample{"field_prefix", "测点前缀", "", "input", 21},
		"field_suffix":    ConfigSample{"field_suffix", "测点后缀", "", "input", 22},
	},
	"modbus": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "modbus", "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "modbus", "input", 2},
		"address":       ConfigSample{"address", "数据源地址", "10.211.55.4:502", "input", 3},
		"slave_id":      ConfigSample{"slave_id", "从站地址", 1, "input", 4},
		"interval":      ConfigSample{"interval", "采集周期", "3s", "combine", 20},
		"field_prefix":  ConfigSample{"field_prefix", "测点前缀", "", "input", 21},
		"field_suffix":  ConfigSample{"field_suffix", "测点后缀", "", "input", 22},
	},
	"fake": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "fake", "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "fake", "input", 2},
		"interval":      ConfigSample{"interval", "采集周期", "3s", "combine", 20},
		"field_prefix":  ConfigSample{"field_prefix", "测点前缀", "", "input", 21},
		"field_suffix":  ConfigSample{"field_suffix", "测点后缀", "", "input", 22},
	},
	"s7": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "s7", "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "s7", "input", 2},
		"address":       ConfigSample{"address", "数据源地址", "192.168.0.168", "input", 3},
		"rack":          ConfigSample{"rack", "机架号", 0, "input", 4},
		"slot":          ConfigSample{"slot", "槽号", 1, "input", 5},
		"interval":      ConfigSample{"interval", "采集周期", "3s", "combine", 20},
		"field_prefix":  ConfigSample{"field_prefix", "测点前缀", "", "input", 21},
		"field_suffix":  ConfigSample{"field_suffix", "测点后缀", "", "input", 22},
	},
	"http_listener": {
		"plugin_name":    ConfigSample{"plugin_name", "插件名称", "http_listener", "select", 1},
		"listen_address": ConfigSample{"listen_address", "监听地址", "0.0.0.0:19999", "input", 2},
		"max_body_size":  ConfigSample{"max_body_size", "最大消息体大小", 5 * 1024 * 1024, "input", 3},
		"max_line_size":  ConfigSample{"max_line_size", "最大文件行数", 64 * 1024, "input", 4},
		"read_timeout":   ConfigSample{"read_timeout", "读超时时间", "10s", "combine", 5},
		"write_timeout":  ConfigSample{"write_timeout", "写超时时间", "10s", "combine", 6},
		"basic_username": ConfigSample{"basic_username", "认证账户", "", "input", 7},
		"basic_password": ConfigSample{"basic_password", "认证密码", "", "input", 8},
	},
}
var OutputSample = map[string]map[string]ConfigSample{
	"_base": {
		"metric_buffer_limit": ConfigSample{"metric_buffer_limit", "批量上传缓冲区大小", 0, "input", 100},
		"metric_batch_size":   ConfigSample{"metric_batch_size", "测点批量上传数量", 0, "input", 101},
		"created_at":          ConfigSample{"created_at", "创建时间", time.Now().UnixNano() / 1e6, "none", 0},
	},
	"file": {
		"plugin_name": ConfigSample{"plugin_name", "插件名称", "file", "select", 1},
		"files":       ConfigSample{"files", "输出地址", []string{"stdout"}, "multi-input", 2},
	},
	"redis": {
		"plugin_name": ConfigSample{"plugin_name", "插件名称", "redis", "select", 1},
		"url_address": ConfigSample{"url_address", "地址URL", "redis://localhost:6379/0", "input", 2},
	},
	"rabbitmq": {
		"plugin_name": ConfigSample{"plugin_name", "插件名称", "rabbitmq", "select", 1},
		"url_address": ConfigSample{"url_address", "地址URL", "amqp://leaniot:leaniot@localhost:5672/gateway", "input", 2},
	},
}
var ControllerSample = map[string]map[string]ConfigSample{
	"_base": {
		"created_at": ConfigSample{"created_at", "创建时间", time.Now().UnixNano() / 1e6, "none", 0},
	},
	"http": {
		"plugin_name": ConfigSample{"plugin_name", "插件名称", "http", "select", 1},
		"address":     ConfigSample{"address", "监听地址", "0.0.0.0:9999", "input", 2},
	},
}
