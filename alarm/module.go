package alarm

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/thegreatco/viam-alarm/utils"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var Model = resource.NewModel("thegreatco", "alarm", "sensor")
var PrettyName = "Sensor Alarm"

type alarm struct {
	resource.Named
	mu                  sync.RWMutex
	logger              logging.Logger
	cancelCtx           context.Context
	cancelFunc          func()
	monitor             func()
	wg                  sync.WaitGroup
	sensor              sensor.Sensor
	cfg                 *AlarmConfig
	queue               *utils.Queue
	currentRateOfChange float64
	currentValue        *sample
}

type sample struct {
	value float64
	time  time.Time
}

func init() {
	resource.RegisterComponent(sensor.API, Model,
		resource.Registration[sensor.Sensor, *AlarmConfig]{
			Constructor: NewAlarmSensor,
		})
}

func NewAlarmSensor(ctx context.Context, desp resource.Dependencies, conf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	logger.Infof("Starting %s %s", PrettyName, utils.Version)
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	sensor := &alarm{
		Named:      conf.ResourceName().AsNamed(),
		logger:     logger,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
		wg:         sync.WaitGroup{},
	}
	if err := sensor.Reconfigure(ctx, desp, conf); err != nil {
		return nil, err
	}
	return sensor, nil
}

func (r *alarm) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"rateOfChange":       r.currentRateOfChange,
		"currentValue":       r.currentValue,
		"lowLowValueAlarm":   r.currentValue.value < r.cfg.LowLowValue,
		"lowValueAlarm":      r.currentValue.value < r.cfg.LowValue,
		"highValueAlarm":     r.currentValue.value > r.cfg.HighValue,
		"highHighValueAlarm": r.currentValue.value > r.cfg.HighHighValue,
	}, nil
}

func (r *alarm) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Debugf("Closing %s", PrettyName)
	r.cancelFunc()
	r.wg.Wait()
	return nil
}

func (*alarm) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"ok": 1}, nil
}

func (r *alarm) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger.Debugf("Reconfiguring %s", PrettyName)

	r.logger.Debugf("Cancelling previous monitor")
	r.cancelFunc()
	r.logger.Debugf("Waiting for monitor to finish")
	r.wg.Wait()
	r.logger.Debugf("Monitor finished, reconfiguring")

	newConf, err := resource.NativeConfig[*AlarmConfig](conf)
	if err != nil {
		return err
	}

	r.cfg = newConf

	sensorName := sensor.Named(newConf.SensorName)
	untypedSensor, err := deps.Lookup(sensorName)
	if err != nil {
		return err
	}
	r.sensor = untypedSensor.(sensor.Sensor)

	r.queue, err = utils.NewQueue(newConf.SampleSize)
	if err != nil {
		return err
	}

	matchRegex, err := regexp.Compile(newConf.ValueRegex)
	if err != nil {
		return err
	}
	sleepTime := 1000 / r.cfg.PollingFrequencyHz

	r.monitor = func() {
		ctx := context.Background()
		r.wg.Add(1)
		defer r.wg.Done()
		for {
			select {
			case <-r.cancelCtx.Done():
				return
			default:
				readings, err := r.sensor.Readings(ctx, nil)
				if err != nil {
					r.logger.Errorf("Error getting readings from sensor: %s", err)
					break
				}
				rawValue := readings[r.cfg.FieldName]
				if rawValue == nil {
					r.logger.Errorf("Field %s not found in readings", r.cfg.FieldName)
					break
				}
				value, err := parseToFloat(rawValue, matchRegex)
				if err != nil {
					r.logger.Errorf("Error parsing value: %s", err)
					break
				}
				r.currentValue = &sample{value: value, time: time.Now()}
				r.queue.Push(value)
				samples := r.queue.ReadAll()
				if len(samples) < 2 {
					break
				}
				for i := 1; i < len(samples); i++ {
					rateOfChange := (samples[i].(float64) - samples[i-1].(float64)) / (float64(sleepTime) / 1000)
					r.currentRateOfChange = rateOfChange
				}
			}

			select {
			case <-time.After(time.Duration(sleepTime) * time.Millisecond):
				continue
			case <-r.cancelCtx.Done():
				return
			}
		}
	}
	return nil
}

func parseToFloat(rawValue interface{}, matchRegex *regexp.Regexp) (float64, error) {
	var value float64
	switch v := rawValue.(type) {
	case string:
		match := matchRegex.FindString(v)
		if match == "" {
			return 0, fmt.Errorf("value %s does not match regex", v)
		}
		p, err := strconv.ParseFloat(match, 64)
		if err != nil {
			return 0, fmt.Errorf("error parsing value %s as float: %s", match, err)
		}
		value = p
	case int:
		value = float64(v)
	case int32:
		value = float64(v)
	case int64:
		value = float64(v)
	case float32:
		value = float64(v)
	case float64:
		value = float64(v)
	}
	return value, nil
}
