package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

// BridgeIPResponse is the response received from the Hue bridge API re: the IP of the bridge
type BridgeIPResponse struct {
	ID                string `json:"id"`
	InternalIPAddress string `json:"internalipaddress"`
}

// SensorDataResponseState is the response received from the Hue bridge API re: the presence recorded by the sensor
type SensorDataResponseState struct {
	Presence bool `json:"presence"`
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
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func getBridgeAddress() (string, error) {
	body, err := httpGet("https://discovery.meethue.com/")
	if err != nil {
		return "", err
	}
	var res []BridgeIPResponse
	json.Unmarshal([]byte(body), &res)
	if len(res) != 1 {
		return "", errors.New("Could not determine bridge IP, response length != 1")
	}

	return res[0].InternalIPAddress, nil
}

func getSensorPresence(bridgeIPAddress string, username string) (bool, error) {
	body, err := httpGet("http://" + bridgeIPAddress + "/api/" + username + "/sensors/4")
	if err != nil {
		return false, err
	}
	var res SensorDataResponse
	json.Unmarshal([]byte(body), &res)
	return res.State.Presence, nil
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

func play(mpdIPAddress string, arg ...string) error {
	return runMpcCmd(mpdIPAddress, "play")
}

func schedule(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}

func setPlayState(bridgeIPAddress string, username string, mpdIPAddress string, graceTime int) {
	presence, err := getSensorPresence(bridgeIPAddress, username)
	if err != nil {
		log.Println(err)
		return
	}

	if presence && !playing {
		play(mpdIPAddress)
		playing = true
		time.Sleep(time.Duration(graceTime) * time.Second)
	} else if !presence && playing {
		pause(mpdIPAddress)
		playing = false
	}
}

func initMpd(mpdIPAddress string, playlist string, volume string) error {
	err := runMpcCmd(mpdIPAddress, "clear")
	if err != nil {
		return errors.New("Could not clear playlist")
	}
	err = runMpcCmd(mpdIPAddress, "load", playlist)
	if err != nil {
		return fmt.Errorf("Could not load playlist %s :(((", playlist)
	}
	err = runMpcCmd(mpdIPAddress, "repeat")
	if err != nil {
		return errors.New("Could not set repeat")
	}
	err = runMpcCmd(mpdIPAddress, "random")
	if err != nil {
		return errors.New("Could not set random")
	}
	err = runMpcCmd(mpdIPAddress, "volume", volume)
	if err != nil {
		return errors.New("Could not set volume")
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
	log.Fatalln(fmt.Sprintf("Could not find env var %s. Quitting.", key))
	os.Exit(1)
	return ""
}

func main() {
	bridgeIPAddress := getEnvOrDie("HUE_BRIDGE_IP")
	mpdIPAddress := getEnvOrDie("MPD_IP")
	username := getEnvOrDie("HUE_USERNAME")
	playlist := getEnvOrDie("PLAYLIST")
	graceTime, err := strconv.Atoi(getEnv("SLEEP_AFTER_AWAY", "30"))
	if err != nil {
		log.Fatalln(fmt.Sprintf("This is not an int: %s, %v", getEnv("SLEEP_AFTER_AWAY", "30"), err))
	}
	volume := getEnv("VOLUME", "5")
	if strconv.Atoi(volume); err != nil {
		log.Fatalln(fmt.Sprintf("This is not an int: %s, %v", volume, err))
	}

	initMpd(mpdIPAddress, playlist, volume)
	setPlayState(bridgeIPAddress, username, mpdIPAddress, graceTime)

	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		log.Println("Program killed!")

		pause(mpdIPAddress)

		os.Exit(0)
	}()

	ticker := time.NewTicker(1500 * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case _ = <-ticker.C:
				setPlayState(bridgeIPAddress, username, mpdIPAddress, graceTime)
			}
		}
	}()
	select {} // block forever
}
