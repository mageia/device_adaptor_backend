package all

import (
	_ "deviceAdaptor/plugins/inputs/fake"
	//_ "deviceAdaptor/plugins/inputs/ftp"
	_ "deviceAdaptor/plugins/inputs/modbus"
	_ "deviceAdaptor/plugins/inputs/s7"
	//_ "deviceAdaptor/plugins/inputs/sensor_vibration"
	_ "deviceAdaptor/plugins/inputs/http_listener"
	_ "deviceAdaptor/plugins/inputs/opc"
)
