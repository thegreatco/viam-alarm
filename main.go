package main

import (
	rocalarm "github.com/thegreatco/viam-alarm/alarm"
	module_utils "github.com/thegreatco/viam-alarm/utils"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
)

func main() {
	module.ModularMain(module_utils.LoggerName,
		resource.APIModel{API: sensor.API, Model: rocalarm.Model},
	)
}
