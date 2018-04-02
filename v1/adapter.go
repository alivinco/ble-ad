package main

import  "github.com/alivinco/fimpgo"
import (
	log "github.com/Sirupsen/logrus"
	"time"
	"github.com/barnybug/miflora"
	"io/ioutil"
	"encoding/json"
)

type MifloraConfig struct {
	MqttClientIdPrefix string
	MqttServerURI string
	MqttUsername string
	MqttPassword string
	MqttTopicGlobalPrefix string
	AdapterName string // hci0
	RetryCount int
	DeviceAddresses []string // List of MAC addresses
	PoolInterval int // interval in seconds
}

type MiFloraAd struct{
	configPath string
	deviceAddresses []string
	msgTransport *fimpgo.MqttTransport
	config MifloraConfig
	runningRequests map [string]bool
}

func NewMifloraAd( configPath string ) *MiFloraAd  {
	mi := MiFloraAd{configPath:configPath}
	mi.runningRequests = map[string]bool{}
	err := mi.loadConfig()
	if err != nil {
		log.Info("<Ad> Can't load config from  ",mi.configPath)
		log.Info("<Ad> Failed to load config. Loading defaults ")
		mi.config.RetryCount = 10
		mi.config.PoolInterval = 600 // seconds
		mi.config.MqttServerURI = "tcp://localhost:1883"
		mi.config.MqttClientIdPrefix = "inst1"
		mi.config.AdapterName = "hci0"
	}else {

	}
	mi.InitMessagingTransport()

	return &mi
}

func (mg *MiFloraAd) loadConfig() error {
	configFileBody, err := ioutil.ReadFile(mg.configPath)
	if err != nil {
		return err
	}
	for _,addr := range mg.config.DeviceAddresses {
		mg.runningRequests[addr] = false
	}
	err = json.Unmarshal(configFileBody, &mg.config)
	return err
}

func (mg *MiFloraAd) saveConfig() {

}

func (mg *MiFloraAd) InitMessagingTransport() error {
	clientId := mg.config.MqttClientIdPrefix + "ble_ad"
	mg.msgTransport = fimpgo.NewMqttTransport(mg.config.MqttServerURI, clientId, mg.config.MqttUsername, mg.config.MqttPassword, true, 1, 1)
	mg.msgTransport.SetGlobalTopicPrefix(mg.config.MqttTopicGlobalPrefix)
	err := mg.msgTransport.Start()
	log.Info("<Ad> Mqtt transport connected")
	if err != nil {
		log.Error("<Ad> Error connecting to broker : ", err)
	}
	mg.msgTransport.SetMessageHandler(mg.onMqttMessage)
	mg.msgTransport.Subscribe("pt:j1/mt:cmd/rt:ad/rn:ble/ad:1")
//pt:j1/mt:evt/rt:dev/rn:zw/ad:1/sv:meter_elec/ad:59_0
	mg.msgTransport.Subscribe("pt:j1/mt:cmd/rt:dev/rn:ble/ad:1/#")
	return err
}


func (mg *MiFloraAd) Start(){
	go mg.pollDevices()
}

func (mg *MiFloraAd) pollDevices(){
	for {
		for _,addr := range mg.config.DeviceAddresses {
			mg.requestSensorData(addr)
			time.Sleep(1*time.Second)
		}
		time.Sleep(time.Duration(mg.config.PoolInterval)*time.Second)
	}

}

func (mg *MiFloraAd) requestSensorData(addr string) {
	log.Info("Requesting sensor data from :",addr)
	defer func() {
		mg.runningRequests[addr] = false
	}()

	if mg.runningRequests[addr] {
		log.Info("Another request is already running.")
		return
	}
	mg.runningRequests[addr]= true
	log.Info("Reading miflora...")
	dev := miflora.NewMiflora(addr, mg.config.AdapterName)

	firmware, err := dev.ReadFirmware()
	if err == nil {
		mg.reportBatteryLevel(addr,firmware.Battery)
	}

	log.Infof("Firmware: %+v\n", firmware)
	var sensors miflora.Sensors
	for i:=0;i<mg.config.RetryCount;i++ {
		sensors, err = dev.ReadSensors()
		if err == nil {
			break
		}else {
			log.Infof("Failed reading sensors,retrying")
		}
		time.Sleep(1*time.Second)
	}
	if err == nil {
		log.Infof("Reporting sensors: %+v\n", sensors)
		if sensors.Temperature < 100 && sensors.Temperature > -50 {
			mg.reportSensor(addr,"sensor_temp",sensors.Temperature,"C")
			mg.reportSensor(addr,"sensor_lumin",float64(sensors.Light),"Lux")
			//mg.reportSensor(addr,"sensor_moist",float64(sensors.Moisture),"%")
			mg.reportSensor(addr,"sensor_humid",float64(sensors.Moisture),"%")
			mg.reportSensor(addr,"sensor_conduct",float64(sensors.Conductivity),"?")
		}else {
			log.Debug("Temp value is outside allowed values ")
		}

	}
}

func (mg *MiFloraAd) reportSensor(addr string ,service string,value float64,unit string) {
	addr = macToFimpMac(addr)
	fimpAddr := fimpgo.Address{MsgType:fimpgo.MsgTypeEvt,ResourceType:fimpgo.ResourceTypeDevice,ResourceName:"ble",ResourceAddress:"1",ServiceName:service,ServiceAddress:addr}
	fimpMsg := fimpgo.NewMessage("evt.sensor.report",service,fimpgo.VTypeFloat,value,fimpgo.Props{"unit":unit},nil,nil)
	mg.msgTransport.Publish(&fimpAddr,fimpMsg)
}
func (mg *MiFloraAd) reportBatteryLevel(addr string , level byte) {
	service := "battery"
	addr = macToFimpMac(addr)
	fimpAddr := fimpgo.Address{MsgType:fimpgo.MsgTypeEvt,ResourceType:fimpgo.ResourceTypeDevice,ResourceName:"ble",ResourceAddress:"1",ServiceName:service,ServiceAddress:addr}
	fimpMsg := fimpgo.NewMessage("evt.lvl.report",service,fimpgo.VTypeInt,level,fimpgo.Props{},nil,nil)
	mg.msgTransport.Publish(&fimpAddr,fimpMsg)

}

func (mg *MiFloraAd) publishInclusionReport(addr string) {

}

