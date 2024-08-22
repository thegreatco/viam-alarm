package alarm

import "errors"

var ErrMissingSensorName = errors.New("missing sensorName")
var ErrMissingFieldName = errors.New("missing fieldName")
var ErrRateOfChangeMustBePositive = errors.New("rateOfChangePerSecond must be positive")

type AlarmConfig struct {
	SensorName            string  `json:"sensorName"`
	FieldName             string  `json:"fieldName"`
	ValueRegex            string  `json:"valueRegex"`
	RateOfChangePerSecond float64 `json:"rateOfChangePerSecond"`
	PollingFrequencyHz    float64 `json:"pollingFrequencyHz"`
	SampleSize            int     `json:"sampleSize"`
	LowValue              float64 `json:"lowValue"`
	LowLowValue           float64 `json:"lowLowValue"`
	HighValue             float64 `json:"highValue"`
	HighHighValue         float64 `json:"highHighValue"`
	// TODO: Add debounce timer
}

func (cfg *AlarmConfig) Validate(path string) ([]string, error) {
	if cfg.SensorName == "" {
		return nil, ErrMissingSensorName
	}
	if cfg.FieldName == "" {
		return nil, ErrMissingFieldName
	}
	if cfg.RateOfChangePerSecond <= 0 {
		return nil, ErrRateOfChangeMustBePositive
	}
	if cfg.PollingFrequencyHz <= 0 {
		return nil, errors.New("pollingFrequencyHz must be positive")
	}
	if cfg.SampleSize < 2 {
		return nil, errors.New("sampleSize must be at least 2")
	}
	return nil, nil
}
