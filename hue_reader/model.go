package hue_reader

// SensorDataResponseState is the response received from the Hue bridge API re: the light level recorded by the sensor
type SensorDataResponseState struct {
	LightLevel int `json:"lightlevel"`
}

// SensorDataResponse is the response received from the Hue bridge API re: state of the sensor
type SensorDataResponse struct {
	State SensorDataResponseState `json:"state"`
}
