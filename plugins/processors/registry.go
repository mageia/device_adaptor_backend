package processors

import "device_adaptor"

type Creator func() device_adaptor.Processor

var Processors = map[string]Creator{}

func Add(name string, creator Creator) {
	Processors[name] = creator
}
