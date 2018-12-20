package all

import (
	_ "deviceAdaptor/plugins/inputs/eip"
	_ "deviceAdaptor/plugins/inputs/fake"
	_ "deviceAdaptor/plugins/inputs/http_listener"
	_ "deviceAdaptor/plugins/inputs/kj66"
	_ "deviceAdaptor/plugins/inputs/modbus"
	_ "deviceAdaptor/plugins/inputs/opc"
	_ "deviceAdaptor/plugins/inputs/s7"
	_ "deviceAdaptor/plugins/inputs/serial"
	//_ "deviceAdaptor/plugins/inputs/ftp"
)
