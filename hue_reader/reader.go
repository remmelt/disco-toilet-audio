package hue_reader

import (
	"encoding/json"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type HueReader struct {
	bridgeIPAddress string
	username        string
}

func NewHueReader(bridgeIPAddress string, username string) *HueReader {
	return &HueReader{
		bridgeIPAddress: bridgeIPAddress,
		username:        username,
	}
}

func (r *HueReader) GetSensorLightLevel() (int, error) {
	body, err := httpGet("http://" + r.bridgeIPAddress + "/api/" + r.username + "/sensors/5")
	if err != nil {
		return 0, err
	}
	var res SensorDataResponse
	err = json.Unmarshal([]byte(body), &res)
	if err != nil {
		return 0, err
	}
	return res.State.LightLevel, nil
}

func httpGet(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err2 := Body.Close()
		if err2 != nil {
			log.Error("could not close body returned from http get", err2)
		}
	}(response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
