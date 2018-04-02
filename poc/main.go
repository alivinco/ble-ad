//shows how to watch for new devices and list them
package poc

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/emitter"
	"github.com/muka/go-bluetooth/linux"
	"time"
	//"github.com/muka/go-bluetooth/devices"
	"errors"
	adevices "github.com/alivinco/ble-ad/devices"
)

const logLevel = log.DebugLevel
const adapterID = "hci0"

func main() {

	log.SetLevel(logLevel)

	//clean up connection on exit
	defer api.Exit()

	log.Debugf("Reset bluetooth device")
	a := linux.NewBtMgmt(adapterID)
	err := a.Reset()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// Load adapter and device info
	adapterID := "hci0"
	//deviceID := "C4:7C:8D:63:7A:1F" // MI Band 2



	devices, err := api.GetDevices()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	log.Infof("Cached devices:")
	for _, dev := range devices {
		showDeviceInfo(&dev)
	}

	log.Infof("Discovered devices:")
	err = discoverDevices(adapterID)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	//manager, err := api.NewManager()
	//if err != nil {
	//	log.Error(err)
	//	os.Exit(1)
	//}
	//
	//err = manager.RefreshState()
	//if err != nil {
	//	log.Error(err)
	//	os.Exit(1)
	//}

	err = ShowSensorTagInfo(adapterID)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	select {}
}

func discoverDevices(adapterID string) error {

	err := api.StartDiscovery()
	if err != nil {
		return err
	}

	log.Debugf("Started discovery")
	err = api.On("discovery", emitter.NewCallback(func(ev emitter.Event) {
		discoveryEvent := ev.GetData().(api.DiscoveredDeviceEvent)
		dev := discoveryEvent.Device
		showDeviceInfo(dev)

	}))

	return err
}

func showDeviceInfo(dev *api.Device) {
	if dev == nil {
		return
	}
	props, err := dev.GetProperties()
	if err != nil {
		log.Errorf("%s: Failed to get properties: %s", dev.Path, err.Error())
		return
	}
	log.Infof("name=%s addr=%s rssi=%d", props.Name, props.Address, props.RSSI)

	log.Infof("Services resolved = ",props.ServicesResolved)
	log.Infof("Service data = ",props.ServiceData)

}

//ShowSensorTagInfo show info from a sensor tag
func ShowSensorTagInfo(adapterID string) error {

	boo, err := api.AdapterExists(adapterID)
	if err != nil {
		return err
	}
	log.Debugf("AdapterExists: %b", boo)

	//err = api.StartDiscoveryOn(adapterID)
	//if err != nil {
	//	return err
	//}
	// wait a moment for the device to be spawn
	time.Sleep(time.Second)

	devarr, err := api.GetDevices()
	if err != nil {
		return err
	}
	//log.Debug("devarr",devarr[0])
	len := len(devarr)
	log.Debugf("length: %d", len)

	for i := 0; i < len; i++ {
		prop1, err := devarr[i].GetProperties()
		if err != nil {
			log.Fatalf("Cannot load properties of %s: %s", devarr[i].Path, err.Error())
			continue
		}
		log.Debugf("DeviceProperties - ADDRESS: %s", prop1.Address)

		if prop1.Address == "C4:7C:8D:63:33:14"{
			err = ConnectAndFetchSensorDetailAndData(prop1.Address)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

// ConnectAndFetchSensorDetailAndData load an show sensor data
func ConnectAndFetchSensorDetailAndData(tagAddress string) error {

	dev, err := api.GetDeviceByAddress(tagAddress)
	if err != nil {
		return err
	}
	log.Debugf("device (dev): %s", dev)

	if dev == nil {
		return errors.New("device not found")
	}

	if !dev.IsConnected() {
		log.Debug("not connected")
		for i:=0;i<10;i++ {
			err = dev.Connect()
			if err == nil {
				log.Info("Connected")
				break
			}
			log.Info("Reconnecting")
			time.Sleep(1*time.Second)
		}

	} else {
		log.Debug("already connected")
	}
	//
	sensorTag, err := adevices.NewMiflora(dev)
	if err != nil {
		return err
	}
	//
	//name := sensorTag.Temperature.GetName()
	//log.Debugf("sensor name: %s", name)
	//
	//name1 := sensorTag.Humidity.GetName()
	//log.Debugf("sensor name: %s", name1)
	//
	//mpu := sensorTag.Mpu.GetName()
	//log.Debugf("sensor name: %s", mpu)
	//
	//barometric := sensorTag.Barometric.GetName()
	//log.Debugf("sensor name: %s", barometric)
	//

	devInfo, err := sensorTag.DeviceInfo.Read()
	if err != nil {
		return err
	}
	//
	log.Debug("FirmwareVersion: ", devInfo.FirmwareVersion)
	log.Debug("HardwareVersion: ", devInfo.HardwareVersion)
	log.Debug("Manufacturer: ", devInfo.Manufacturer)
	//log.Debug("Model: ", devInfo.Model)
	//
	//err = sensorTag.Temperature.StartNotify()
	//if err != nil {
	//	return err
	//}
	//
	//err = sensorTag.Humidity.StartNotify()
	//if err != nil {
	//	return err
	//}
	//
	//err = sensorTag.Mpu.StartNotify(tagAddress)
	//if err != nil {
	//	return err
	//}
	//
	//err = sensorTag.Barometric.StartNotify(tagAddress)
	//if err != nil {
	//	return err
	//}
	//
	//err = sensorTag.Luxometer.StartNotify(tagAddress)
	//if err != nil {
	//	return err
	//}

	err = dev.On("data", emitter.NewCallback(func(ev emitter.Event) {
		x := ev.GetData().(api.DataEvent)
		log.Debugf("Data: %++v", x)
	}))

	return err
}


