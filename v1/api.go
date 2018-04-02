package main

import (
    "github.com/alivinco/fimpgo/fimptype"
    "github.com/alivinco/fimpgo"
	"github.com/labstack/gommon/log"
	"strings"
	"strconv"
)

func (mg *MiFloraAd) onMqttMessage(topic string, addr *fimpgo.Address, iotMsg *fimpgo.FimpMessage, rawMessage []byte) {
	if addr.ResourceType == fimpgo.ResourceTypeAdapter{
		switch iotMsg.Type {
		case "cmd.thing.get_inclusion_report":
			devAddr,err := iotMsg.GetStringValue()
			if err != nil {
				log.Error("Value is not string")
				return
			}
			for _,a := range mg.config.DeviceAddresses {

				if a == fimpMacToMac(devAddr) {
					mg.SendInclusionReport(devAddr)
					break
				}
			}
		case "cmd.network.get_all_nodes":
			mg.SendDeviceListReport()


		}
	}else if addr.ResourceType == fimpgo.ResourceTypeDevice {
		switch iotMsg.Type {
		case "cmd.sensor.get_report":
			if addr.ServiceAddress == "" {
				log.Error("Address is empty")
				return
			}
			mg.requestSensorData(fimpMacToMac(addr.ServiceAddress))
		}
	}
}

type DeviceListItem struct {
	Address string `json:"address"`
}

func (mg *MiFloraAd) SendDeviceListReport() {
	var listOfDevices []DeviceListItem
	for _,dev := range mg.config.DeviceAddresses {
		listOfDevices = append(listOfDevices,DeviceListItem{Address:dev})
	}

	msg := fimpgo.NewMessage("evt.network.all_nodes_report", "ble","object", listOfDevices, nil,nil,nil)
	addrString := "pt:j1/mt:evt/rt:ad/rn:ble/ad:1"
	fimpAddr, _ := fimpgo.NewAddressFromString(addrString)
	mg.msgTransport.Publish(fimpAddr,msg)

}

func macToFimpMac(addr string)string {
	return strings.Replace(addr,":","-",-1)
}

func fimpMacToMac(addr string)string {
	return strings.Replace(addr,"-",":",-1)
}

func (mg *MiFloraAd) SendInclusionReport(addr string) {

	report := fimptype.ThingInclusionReport{}
	report.Type = "ble"
	report.Address = addr
	report.Alias = "Flower care"
	report.CommTechnology = "ble"
	report.PowerSource = "battery"
	report.WakeUpInterval = strconv.Itoa(mg.config.PoolInterval)
	report.ProductName = "Flower care"
	report.ProductHash = "flower_care_1"
	report.SwVersion = "1.0"
	report.Groups = []string{}
	report.ProductId = "flower_care"
	report.ManufacturerId = "mi"
	report.Security = "tls"
	report.Groups = []string{"ch_0"}


	var services []fimptype.Service
    // <Service>
	service := fimptype.Service{}
	service.Name = "sensor_temp"
	service.Alias = ""
	service.Enabled = true
	service.Address = "/rt:dev/rn:"+report.Type+"/ad:1/sv:"+service.Name+"/ad:"+addr
	service.Groups = []string{"ch_0"}
	service.Interfaces = []fimptype.Interface{}
	service.Props = map[string]interface{}{"sup_units":[]string{"C"}}
	service.Tags = []string{}

	intf := fimptype.Interface{}
	intf.Type = "out"
	intf.MsgType = "evt.sensor.report"
	intf.ValueType = fimpgo.VTypeFloat
	intf.Version = "1"
	service.Interfaces = append(service.Interfaces,intf)

	intf = fimptype.Interface{}
	intf.Type = "in"
	intf.MsgType = "cmd.sensor.get_report"
	intf.ValueType = fimpgo.VTypeString
	intf.Version = "1"
	service.Interfaces = append(service.Interfaces,intf)


	services = append(services,service)
	// </Service>


	// <Service>
	service = fimptype.Service{}
	service.Name = "sensor_lumin"
	service.Alias = ""
	service.Enabled = true
	service.Address = "/rt:dev/rn:"+report.Type+"/ad:1/sv:"+service.Name+"/ad:"+addr
	service.Groups = []string{"ch_0"}
	service.Interfaces = []fimptype.Interface{}
	service.Props = map[string]interface{}{"sup_units":[]string{"Lux"}}
	service.Tags = []string{}

	intf = fimptype.Interface{}
	intf.Type = "out"
	intf.MsgType = "evt.sensor.report"
	intf.ValueType = fimpgo.VTypeFloat
	intf.Version = "1"
	service.Interfaces = append(service.Interfaces,intf)

	intf = fimptype.Interface{}
	intf.Type = "in"
	intf.MsgType = "cmd.sensor.get_report"
	intf.ValueType = fimpgo.VTypeString
	intf.Version = "1"
	service.Interfaces = append(service.Interfaces,intf)


	services = append(services,service)
	// </Service>

	// <Service>
	service = fimptype.Service{}
	service.Name = "sensor_humid"
	service.Alias = ""
	service.Enabled = true
	service.Address = "/rt:dev/rn:"+report.Type+"/ad:1/sv:"+service.Name+"/ad:"+addr
	service.Groups = []string{"ch_0"}
	service.Interfaces = []fimptype.Interface{}
	service.Props = map[string]interface{}{"sup_units":[]string{"%"}}
	service.Tags = []string{}

	intf = fimptype.Interface{}
	intf.Type = "out"
	intf.MsgType = "evt.sensor.report"
	intf.ValueType = fimpgo.VTypeFloat
	intf.Version = "1"
	service.Interfaces = append(service.Interfaces,intf)

	intf = fimptype.Interface{}
	intf.Type = "in"
	intf.MsgType = "cmd.sensor.get_report"
	intf.ValueType = fimpgo.VTypeString
	intf.Version = "1"
	service.Interfaces = append(service.Interfaces,intf)


	services = append(services,service)
	// </Service>

	// <Service>
	service = fimptype.Service{}
	service.Name = "sensor_conduct"
	service.Alias = ""
	service.Enabled = true
	service.Address = "/rt:dev/rn:"+report.Type+"/ad:1/sv:"+service.Name+"/ad:"+addr
	service.Groups = []string{"ch_0"}
	service.Interfaces = []fimptype.Interface{}
	service.Props = map[string]interface{}{"sup_units":[]string{"?"}}

	intf = fimptype.Interface{}
	intf.Type = "out"
	intf.MsgType = "evt.sensor.report"
	intf.ValueType = fimpgo.VTypeFloat
	intf.Version = "1"
	service.Interfaces = append(service.Interfaces,intf)

	intf = fimptype.Interface{}
	intf.Type = "in"
	intf.MsgType = "cmd.sensor.get_report"
	intf.ValueType = fimpgo.VTypeString
	intf.Version = "1"
	service.Interfaces = append(service.Interfaces,intf)

	service.Tags = []string{}
	services = append(services,service)
	// </Service>

	// <Service>
	service = fimptype.Service{}
	service.Name = "battery"
	service.Alias = ""
	service.Enabled = true
	service.Address = "/rt:dev/rn:"+report.Type+"/ad:1/sv:"+service.Name+"/ad:"+addr
	service.Groups = []string{"ch_0"}
	service.Interfaces = []fimptype.Interface{}
	service.Props = map[string]interface{}{}

	intf = fimptype.Interface{}
	intf.Type = "out"
	intf.MsgType = "evt.lvl.report"
	intf.ValueType = fimpgo.VTypeFloat
	intf.Version = "1"
	service.Interfaces = append(service.Interfaces,intf)

	intf = fimptype.Interface{}
	intf.Type = "in"
	intf.MsgType = "cmd.lvl.get_report"
	intf.ValueType = fimpgo.VTypeString
	intf.Version = "1"
	service.Interfaces = append(service.Interfaces,intf)

	service.Tags = []string{}
	services = append(services,service)
	// </Service>

	report.Services = services

	msg := fimpgo.NewMessage("evt.thing.inclusion_report", report.Type,"object", report, nil,nil,nil)
	addrString := "pt:j1/mt:evt/rt:ad/rn:"+report.Type+"/ad:1"
	fimpAddr, _ := fimpgo.NewAddressFromString(addrString)
	mg.msgTransport.Publish(fimpAddr,msg)
}