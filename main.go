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

type BridgeIpResponse struct {
	Id                string `json:"id"`
	InternalIpAddress string `json:"internalipaddress"`
}
type SensorDataResponseState struct {
	Presence bool `json:"presence"`
}
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
	var res []BridgeIpResponse
	json.Unmarshal([]byte(body), &res)
	if len(res) != 1 {
		return "", errors.New("Could not determine bridge IP, response length != 1")
	}

	return res[0].InternalIpAddress, nil
}

func getSensorPresence(bridgeIpAddress string, username string) (bool, error) {
	body, err := httpGet("http://" + bridgeIpAddress + "/api/" + username + "/sensors/4")
	if err != nil {
		return false, err
	}
	var res SensorDataResponse
	json.Unmarshal([]byte(body), &res)
	return res.State.Presence, nil
}

func runMpcCmd(mpdIpAddress string, arg ...string) error {
	arg = append([]string{"-h", "192.168.178.89"}, arg...)
	log.Println(arg)
	cmd := exec.Command("mpc", arg...)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func pause(mpdIpAddress string) error {
	return runMpcCmd(mpdIpAddress, "pause")
}

func play(mpdIpAddress string, arg ...string) error {
	return runMpcCmd(mpdIpAddress, "play")
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

func setPlayState(bridgeIpAddress string, username string, mpdIpAddress string, graceTime int) {
	presence, err := getSensorPresence(bridgeIpAddress, username)
	if err != nil {
		log.Println(err)
	}

	if presence && !playing {
		play(mpdIpAddress)
		playing = true
		time.Sleep(30 * time.Second)
	} else if !presence && playing {
		pause(mpdIpAddress)
		playing = false
	}
}

func initMpd(mpdIpAddress string, playlist string) error {
	err := runMpcCmd(mpdIpAddress, "clear")
	if err != nil {
		return errors.New("Could not clear playlist")
	}
	err = runMpcCmd(mpdIpAddress, "load", playlist)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not load playlist %s :(((", playlist))
	}
	err = runMpcCmd(mpdIpAddress, "repeat")
	if err != nil {
		return errors.New("Could not set repeat")
	}
	err = runMpcCmd(mpdIpAddress, "volume", "15")
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
	bridgeIpAddress := getEnvOrDie("HUE_BRIDGE_IP")
	mpdIpAddress := getEnvOrDie("MPD_IP")
	username := getEnvOrDie("HUE_USERNAME")
	playlist := getEnvOrDie("PLAYLIST")
	graceTime, err := strconv.Atoi(getEnv("SLEEP_AFTER_AWAY", "30"))
	if err != nil {
		log.Fatalln(fmt.Sprintf("This is not an int: %s, %v", getEnv("SLEEP_AFTER_AWAY", "30"), err))
	}

	initMpd(mpdIpAddress, playlist)
	setPlayState(bridgeIpAddress, username, mpdIpAddress, graceTime)

	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		log.Println("Program killed !")

		pause(mpdIpAddress)

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
				setPlayState(bridgeIpAddress, username, mpdIpAddress, graceTime)
			}
		}
	}()
	select {} // block forever
}
