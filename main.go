package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"time"
)

var playing = false

// SensorDataResponseState is the response received from the Hue bridge API re: the light level recorded by the sensor
type SensorDataResponseState struct {
	LightLevel int `json:"lightlevel"`
}

// SensorDataResponse is the response received from the Hue bridge API re: state of the sensor
type SensorDataResponse struct {
	State SensorDataResponseState `json:"state"`
}

func httpGet(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func getSensorLightLevel(bridgeIPAddress string, username string) (int, error) {
	body, err := httpGet("http://" + bridgeIPAddress + "/api/" + username + "/sensors/5")
	if err != nil {
		return -1, err
	}
	var res SensorDataResponse
	err = json.Unmarshal([]byte(body), &res)
	if err != nil {
		return -1, err
	}
	return res.State.LightLevel, nil
}

func runMpcCmd(mpdIPAddress string, arg ...string) error {
	arg = append([]string{"-h", mpdIPAddress}, arg...)
	log.Println(arg)
	cmd := exec.Command("mpc", arg...)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func pause(mpdIPAddress string) error {
	return runMpcCmd(mpdIPAddress, "pause")
}

func play(mpdIPAddress string) error {
	return runMpcCmd(mpdIPAddress, "play")
}

func setPlayState(bridgeIPAddress string, username string, mpdIPAddress string, dayStart time.Time, dayEnd time.Time, loc *time.Location) {
	lightLevel, err := getSensorLightLevel(bridgeIPAddress, username)
	if err != nil {
		log.Println(err)
		return
	}

	shouldTurnOn := shouldTurnOn(lightLevel, dayStart, dayEnd, loc)

	if shouldTurnOn && !playing {
		err := play(mpdIPAddress)
		if err != nil {
			return
		}
		playing = true
	} else if !shouldTurnOn && playing {
		err := pause(mpdIPAddress)
		if err != nil {
			return
		}
		playing = false
	}
}

func shouldTurnOn(lightLevel int, dayStart time.Time, dayEnd time.Time, loc *time.Location) bool {
	on := false
	if lightLevel > 1000 {
		on = true
	}

	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), dayStart.Hour(), dayStart.Minute(), 0, 0, loc)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), dayEnd.Hour(), dayEnd.Minute(), 0, 0, loc)

	if startOfDay.After(now) || endOfDay.Before(now) {
		on = false
	}

	return on
}

func initMpd(mpdIPAddress string, playlist string, volume string) error {
	err := runMpcCmd(mpdIPAddress, "clear")
	if err != nil {
		return errors.New("could not clear playlist")
	}
	err = runMpcCmd(mpdIPAddress, "load", playlist)
	if err != nil {
		return fmt.Errorf("could not load playlist %s :(((", playlist)
	}
	err = runMpcCmd(mpdIPAddress, "repeat")
	if err != nil {
		return errors.New("could not set repeat")
	}
	err = runMpcCmd(mpdIPAddress, "random")
	if err != nil {
		return errors.New("could not set random")
	}
	err = runMpcCmd(mpdIPAddress, "volume", volume)
	if err != nil {
		return errors.New("could not set volume")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvOrDie(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Fatalln(fmt.Sprintf("could not find env var %s", key))
	return ""
}

func main() {
	bridgeIPAddress := getEnvOrDie("HUE_BRIDGE_IP")
	mpdIPAddress := getEnvOrDie("MPD_IP")
	username := getEnvOrDie("HUE_USERNAME")
	playlist := getEnvOrDie("PLAYLIST")
	dayStartEnv := getEnvOrDie("DAY_START")
	dayEndEnv := getEnvOrDie("DAY_END")

	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		log.Fatalln(fmt.Sprintf("Could not parse location, %v", err))
	}

	dayStart, err := time.ParseInLocation("15:04", dayStartEnv, loc)
	if err != nil {
		log.Fatalln(fmt.Sprintf("DAY_START is not a time: %s, %v", dayStartEnv, err))
	}

	dayEnd, _ := time.ParseInLocation("15:04", dayEndEnv, loc)
	if err != nil {
		log.Fatalln(fmt.Sprintf("DAY_END is not a time: %s, %v", dayEndEnv, err))
	}

	volume := getEnv("VOLUME", "2")
	_, err = strconv.Atoi(volume)
	if err != nil {
		log.Fatalln(fmt.Sprintf("VOLUME is not an int: %s, %v", volume, err))
	}

	err = initMpd(mpdIPAddress, playlist, volume)
	if err != nil {
		return
	}
	setPlayState(bridgeIPAddress, username, mpdIPAddress, dayStart, dayEnd, loc)

	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		log.Println("Program killed!")

		err := pause(mpdIPAddress)
		if err != nil {
			return
		}

		os.Exit(0)
	}()

	ticker := time.NewTicker(2 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case _ = <-ticker.C:
				setPlayState(bridgeIPAddress, username, mpdIPAddress, dayStart, dayEnd, loc)
			}
		}
	}()
	select {} // block forever
}
