package devices

import (
"errors"
"fmt"
"time"

log "github.com/Sirupsen/logrus"
"github.com/godbus/dbus"
"github.com/muka/go-bluetooth/api"
"github.com/muka/go-bluetooth/bluez/profile"
"github.com/muka/go-bluetooth/emitter"
)

// DefaultRetry times
const DefaultRetry = 3

// DefaultRetryWait in millis
const DefaultRetryWait = 500

var dataChannel chan dbus.Signal

//DEBU[0001] device (dev): &{/org/bluez/hci0/dev_C4_7C_8D_63_33_14 %!s(*profile.Device1Properties=&{[]
// [00001204-0000-1000-8000-00805f9b34fb 00001206-0000-1000-8000-00805f9b34fb 00001800-0000-1000-8000-00805f9b34fb
// 00001801-0000-1000-8000-00805f9b34fb 0000fe95-0000-1000-8000-00805f9b34fb 0000fef5-0000-1000-8000-00805f9b34fb]
// false false false false false false map[0000fe95-0000-1000-8000-00805f9b34fb:{{ay} [49 2 152 0 8 20 51 99 141 124 196 13]}]
// map[] 0 0 /org/bluez/hci0 C4:7C:8D:63:33:14  Flower care   Flower care 0 0}) %!s(*profile.Device1=&{0x120357f0 0x11f7e8c0}) map[]}

var MifloraUUIDs = map[string]string{

	"TemperatureData":   "AA01",
	"TemperatureConfig": "AA02",
	"TemperaturePeriod": "AA03",

	"AccelerometerData":   "AA11",
	"AccelerometerConfig": "AA12",
	"AccelerometerPeriod": "AA13",

	"HumidityData":   "AA21",
	"HumidityConfig": "AA22",
	"HumidityPeriod": "AA23",

	"MagnetometerData":   "AA31",
	"MagnetometerConfig": "AA32",
	"MagnetometerPeriod": "AA33",

	"BarometerData":        "AA41",
	"BarometerConfig":      "AA42",
	"BarometerPeriod":      "AA44",
	"BarometerCalibration": "AA43",

	"GyroscopeData":   "AA51",
	"GyroscopeConfig": "AA52",
	"GyroscopePeriod": "AA53",

	"TestData":   "AA61",
	"TestConfig": "AA62",

	"ConnectionParams":        "CCC1",
	"ConnectionReqConnParams": "CCC2",
	"ConnectionDisconnReq":    "CCC3",

	"OADImageIdentify": "FFC1",
	"OADImageBlock":    "FFC2",

	"MPU9250_DATA_UUID":   "AA81",
	"MPU9250_CONFIG_UUID": "AA82",
	"MPU9250_PERIOD_UUID": "AA83",

	"LUXOMETER_CONFIG_UUID": "aa72",
	"LUXOMETER_DATA_UUID":   "aa71",
	"LUXOMETER_PERIOD_UUID": "aa73",

	"DEVICE_INFORMATION_UUID": "180A",
	"SYSTEM_ID_UUID":          "2A23",
	"MODEL_NUMBER_UUID":       "2A24",
	"SERIAL_NUMBER_UUID":      "2A25",
	"FIRMWARE_REVISION_UUID":  "2A26",
	"HARDWARE_REVISION_UUID":  "2A27",
	"SOFTWARE_REVISION_UUID":  "2A28",
	"MANUFACTURER_NAME_UUID":  "2A29",
}

//MifloraDataEvent contains MifloraSpecific data structure
type MifloraDataEvent struct {
	Device     *api.Device
	SensorType string

	AmbientTempValue interface{}
	AmbientTempUnit  string

	ObjectTempValue interface{}
	ObjectTempUnit  string

	SensorID string

	BarometericPressureValue interface{}
	BarometericPressureUnit  string

	BarometericTempValue interface{}
	BarometericTempUnit  string

	HumidityValue interface{}
	HumidityUnit  string

	HumidityTempValue interface{}
	HumidityTempUnit  string

	MpuGyroscopeValue interface{}
	MpuGyroscopeUnit  string

	MpuAccelerometerValue interface{}
	MpuAccelerometerUnit  string

	MpuMagnetometerValue interface{}
	MpuMagnetometerUnit  string

	LuxometerValue interface{}
	LuxometerUnit  string

	FirmwareVersion string
	HardwareVersion string
	Manufacturer    string
	Model           string
}

//Period =[Input*10]ms,(lowerlimit 300 ms, max 2500ms),default 1000 ms
const (
	TemperaturePeriodHigh   = 0x32  // 500 ms,
	TemperaturePeriodMedium = 0x64  // 1000 ms,
	TemperaturePeriodLow    = 0x128 // 2000 ms,
)

func getUUID(name string) (string, error) {
	if MifloraUUIDs[name] == "" {
		return "", fmt.Errorf("Not found %s", name)
	}
	return fmt.Sprintf("F000%s-0451-4000-B000-000000000000", MifloraUUIDs[name]), nil
}

func getDeviceInfoUUID(name string) string {
	if MifloraUUIDs[name] == "" {
		panic("Not found " + name)
	}
	return "0000" + MifloraUUIDs[name] + "-0000-1000-8000-00805F9B34FB"
}

//retryCall n. times, sleep millis, callback
func retryCall(times int, sleep int64, fn func() (interface{}, error)) (intf interface{}, err error) {
	for i := 0; i < times; i++ {
		intf, err = fn()
		if err == nil {
			return intf, nil
		}
		time.Sleep(time.Millisecond * time.Duration(sleep))
	}
	return nil, err
}

//NewMiflora creates a new Miflora instance
func NewMiflora(d *api.Device) (*Miflora, error) {

	s := new(Miflora)

	var connect = func(dev *api.Device) error {
		if !dev.IsConnected() {
			err := dev.Connect()
			if err != nil {
				return err
			}
		}
		return nil
	}

	err := d.On("changed", emitter.NewCallback(func(ev emitter.Event) {
		changed := ev.GetData().(api.PropertyChangedEvent)
		if changed.Field == "Connected" {
			conn := changed.Value.(bool)
			if !conn {

				// TODO clean up properly

				if dataChannel != nil {
					close(dataChannel)
				}

			}
		}
	}))

	if err != nil {
		return nil, err
	}

	err = connect(d)
	if err != nil {
		log.Warning("Miflora connection failed: %v", err)
		return nil, err
	}

	s.Device = d


	//initiating things for reading device info of  Miflora...(getting firmware,hardware,manufacturer,model char...).....

	devInformation, err := newDeviceInfo(s)
	if err != nil {
		return nil, err
	}
	s.DeviceInfo = devInformation

	return s, nil

}

//Miflora a Miflora object representation
type Miflora struct {
	*api.Device
	DeviceInfo  MifloraDeviceInfo
}

//Sensor generic sensor interface
type Sensor interface {
	GetName() string
	IsEnabled() (bool, error)
	Enable() error
	Disable() error
}

func newDeviceInfo(tag *Miflora) (MifloraDeviceInfo, error) {

	dev := tag.Device

	DeviceFirmwareUUID := getDeviceInfoUUID("FIRMWARE_REVISION_UUID")
	DeviceHardwareUUID := getDeviceInfoUUID("HARDWARE_REVISION_UUID")
	DeviceManufacturerUUID := getDeviceInfoUUID("MANUFACTURER_NAME_UUID")
	DeviceModelUUID := getDeviceInfoUUID("MODEL_NUMBER_UUID")

	var loadChars func() (MifloraDeviceInfo, error)

	loadChars = func() (MifloraDeviceInfo, error) {

		firmwareInfo, err := dev.GetCharByUUID(DeviceFirmwareUUID)
		if err != nil {
			return MifloraDeviceInfo{}, err
		}
		if firmwareInfo == nil {
			return MifloraDeviceInfo{}, errors.New("Cannot find DeviceFirmwareUUID characteristic " + DeviceFirmwareUUID)
		}

		hardwareInfo, err := dev.GetCharByUUID(DeviceHardwareUUID)
		if err != nil {
			return MifloraDeviceInfo{}, err
		}
		if hardwareInfo == nil {
			return MifloraDeviceInfo{}, errors.New("Cannot find DeviceHardwareUUID characteristic " + DeviceHardwareUUID)
		}

		manufacturerInfo, err := dev.GetCharByUUID(DeviceManufacturerUUID)
		if err != nil {
			return MifloraDeviceInfo{}, err
		}
		if manufacturerInfo == nil {
			return MifloraDeviceInfo{}, errors.New("Cannot find DeviceManufacturerUUID characteristic " + DeviceManufacturerUUID)
		}

		modelInfo, err := dev.GetCharByUUID(DeviceModelUUID)
		if err != nil {
			return MifloraDeviceInfo{}, err
		}
		if modelInfo == nil {
			return MifloraDeviceInfo{}, errors.New("Cannot find DeviceModelUUID characteristic " + DeviceModelUUID)
		}

		return MifloraDeviceInfo{tag, modelInfo, manufacturerInfo, hardwareInfo, firmwareInfo}, err
	}

	return loadChars()
}

//MifloraDeviceInfo Miflora structure
type MifloraDeviceInfo struct {
	tag              *Miflora
	firmwareInfo     *profile.GattCharacteristic1
	hardwareInfo     *profile.GattCharacteristic1
	manufacturerInfo *profile.GattCharacteristic1
	modelInfo        *profile.GattCharacteristic1
}

//Read device info from Miflora
func (s *MifloraDeviceInfo) Read() (*MifloraDataEvent, error) {

	options1 := make(map[string]dbus.Variant)
	fw, err := s.firmwareInfo.ReadValue(options1)
	if err != nil {
		return nil, err
	}
	options2 := make(map[string]dbus.Variant)
	hw, err := s.hardwareInfo.ReadValue(options2)
	if err != nil {
		return nil, err
	}
	options3 := make(map[string]dbus.Variant)
	manufacturer, err := s.manufacturerInfo.ReadValue(options3)
	if err != nil {
		return nil, err
	}
	options4 := make(map[string]dbus.Variant)
	model, err := s.modelInfo.ReadValue(options4)
	if err != nil {
		return nil, err
	}
	dataEvent := MifloraDataEvent{
		FirmwareVersion: string(fw),
		HardwareVersion: string(hw),
		Manufacturer:    string(manufacturer),
		Model:           string(model),
	}
	return &dataEvent, err
}
