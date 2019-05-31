package all

import (
	_ "device_adaptor/plugins/outputs/amqp"
	_ "device_adaptor/plugins/outputs/file"
	_ "device_adaptor/plugins/outputs/mqtt"
	_ "device_adaptor/plugins/outputs/redis"
)
