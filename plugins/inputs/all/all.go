package all

import (
	_ "device_adaptor/plugins/inputs/eip"
	_ "device_adaptor/plugins/inputs/fake"
	_ "device_adaptor/plugins/inputs/ftp"
	_ "device_adaptor/plugins/inputs/http_listener"
	_ "device_adaptor/plugins/inputs/kj66"
	_ "device_adaptor/plugins/inputs/modbus"
	_ "device_adaptor/plugins/inputs/s7"
	_ "device_adaptor/plugins/inputs/serial"
	_ "device_adaptor/plugins/inputs/snmp"
	//_ "device_adaptor/plugins/inputs/ws"
)
