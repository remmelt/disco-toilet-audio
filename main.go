package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/remmelt/disco-toilet-audio/hue_reader"
	"github.com/remmelt/disco-toilet-audio/mopidy_player"
	log "github.com/sirupsen/logrus"
)

func shouldTurnOn(lightLevel int, dayStart time.Time, dayEnd time.Time, loc *time.Location) bool {
	on := false
	if lightLevel > 1000 {
		on = true
	}
	msg := fmt.Sprintf("found lightLevel: %d", lightLevel)

	if on {
		now := time.Now().In(loc)
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), dayStart.Hour(), dayStart.Minute(), 0, 0, loc)
		endOfDay := time.Date(now.Year(), now.Month(), now.Day(), dayEnd.Hour(), dayEnd.Minute(), 0, 0, loc)

		if startOfDay.After(now) || endOfDay.Before(now) {
			on = false
			msg += fmt.Sprintf(" but not playing because %s not within range", now)
			log.Println(msg)
		}
	}

	//log.WithField("shouldTurnOn", on).WithField("lightLevel", lightLevel).Info("state")

	return on
}

func checkLevel(
	hueReader *hue_reader.HueReader, mopidyClient *mopidy_player.MopidyPlayer, dayStart, dayEnd time.Time,
	loc *time.Location,
) error {
	lightLevel, err := hueReader.GetSensorLightLevel()
	if err != nil {
		return err
	}

	turnOn := shouldTurnOn(lightLevel, dayStart, dayEnd, loc)

	isPlaying, err := mopidyClient.IsPlaying()
	if err != nil {
		return err
	}
	if turnOn && !isPlaying {
		log.Info("playing")
		if err = mopidyClient.Next(); err != nil {
			return err
		}
		if err = mopidyClient.Play(); err != nil {
			return err
		}
	}
	if !turnOn && isPlaying {
		log.Info("pausing")
		if err = mopidyClient.Pause(); err != nil {
			return err
		}
	}

	return nil
}

func createMopidyClient() (*mopidy_player.MopidyPlayer, error) {
	mopidyIPAddress := mustGetEnv("MOPIDY_HTTP_IP")

	v := getEnv("VOLUME", "2")
	volume, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalln(fmt.Sprintf("VOLUME is not an int: %s, %v", v, err))
	}

	return mopidy_player.NewMopidyHttpClient(mopidyIPAddress, volume)
}

func createHueClient() *hue_reader.HueReader {
	bridgeIPAddress := mustGetEnv("HUE_BRIDGE_IP")
	username := mustGetEnv("HUE_USERNAME")

	return hue_reader.NewHueReader(bridgeIPAddress, username)
}

func main() {
	log.Println("starting disco toilet (version tidal/mopidy-http)")

	mopidyClient, err := createMopidyClient()
	if err != nil {
		log.Fatalln("could not init mopidy http client", err)
	}

	hueClient := createHueClient()

	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		log.Fatalln(fmt.Sprintf("could not parse location, %v", err))
	}

	dayStartEnv := mustGetEnv("DAY_START")
	dayEndEnv := mustGetEnv("DAY_END")
	dayStart, err := time.ParseInLocation("15:04", dayStartEnv, loc)
	if err != nil {
		log.Fatalln(fmt.Sprintf("DAY_START is not a time: %s, %v", dayStartEnv, err))
	}

	dayEnd, err := time.ParseInLocation("15:04", dayEndEnv, loc)
	if err != nil {
		log.Fatalln(fmt.Sprintf("DAY_END is not a time: %s, %v", dayEndEnv, err))
	}

	go signalLoop(mopidyClient)
	go tickerLoop(hueClient, mopidyClient, dayStart, dayEnd, loc)

	select {} // block forever
}

func signalLoop(mopidyClient *mopidy_player.MopidyPlayer) {
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	log.Println("program killed")

	if err := mopidyClient.Pause(); err != nil {
		log.Println("Could not pause mopidyClient", err)
	}

	os.Exit(0)
}

func tickerLoop(
	hueClient *hue_reader.HueReader, mopidyClient *mopidy_player.MopidyPlayer, dayStart, dayEnd time.Time,
	loc *time.Location,
) {
	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := checkLevel(hueClient, mopidyClient, dayStart, dayEnd, loc); err != nil {
				log.Println("could not check level", err)
			}
		}
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func mustGetEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Fatalln(fmt.Sprintf("could not find env var %s", key))
	return ""
}
