package mopidy_player

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/sosedoff/musicbot/mopidy"
)

type MopidyPlayer struct {
	address string
	volume  int

	client *mopidy.Client
}

func NewMopidyHttpClient(address string, volume int) (*MopidyPlayer, error) {
	mClient := mopidy.New(address)
	err := mClient.Connect()

	if err != nil {
		return nil, err
	}

	m := &MopidyPlayer{
		address: address,
		volume:  volume,
		client:  mClient,
	}

	if err = m.prepPlayback(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *MopidyPlayer) prepPlayback() error {
	if err := m.setVolume(); err != nil {
		return err
	}
	if err := m.Pause(); err != nil {
		return err
	}
	if err := m.Next(); err != nil {
		return err
	}
	if err := m.setRepeat(); err != nil {
		return err
	}
	if err := m.setRandom(); err != nil {
		return err
	}

	m.logState()
	m.logVolume()
	return nil
}

func (m *MopidyPlayer) logVolume() {
	if volume, err := m.getVolume(); err != nil {
		log.Warning("Error getting volume: ", err)
	} else {
		log.Info("Volume: ", volume)
	}
}

func (m *MopidyPlayer) logState() {
	if state, err := m.getState(); err != nil {
		log.Warning("Error getting state: ", err)
	} else {
		log.Info("State: ", state)
	}
}

func (m *MopidyPlayer) getState() (State, error) {
	state, err := m.client.State()
	if err != nil {
		return UNKNOWN, err
	}
	switch state {
	case "paused":
		return PAUSED, nil
	case "playing":
		return PLAYING, nil
	case "stopped":
		return STOPPED, nil
	}
	return UNKNOWN, fmt.Errorf("unknown state: %s", state)
}

func (m *MopidyPlayer) Next() error {
	return m.client.PlayNextTrack()
}

func (m *MopidyPlayer) Pause() error {
	_, err := m.client.Call("core.playback.pause", nil)
	return err
}

func (m *MopidyPlayer) Play() error {
	if err := m.setVolume(); err != nil {
		return err
	}
	_, err := m.client.Call("core.playback.play", nil)
	return err
}

func (m *MopidyPlayer) IsPlaying() (bool, error) {
	state, err := m.getState()
	if err != nil {
		return false, err
	}
	return state == PLAYING, nil
}

func (m *MopidyPlayer) getVolume() (int, error) {
	data, err := m.client.Call("core.mixer.get_volume", nil)
	if err != nil {
		return 0, err
	}

	resp := data.(*mopidy.BasicResponse)
	return int(resp.Result.(float64)), nil
}

func (m *MopidyPlayer) setVolume() error {
	params := map[string]int{"volume": m.volume}
	_, err := m.client.Call("core.mixer.set_volume", params)
	return err
}

func (m *MopidyPlayer) setRepeat() error {
	if _, err := m.client.Call("core.tracklist.set_repeat", []bool{true}); err != nil {
		return err
	}
	r, err := m.client.Call("core.tracklist.get_repeat", nil)
	if err != nil {
		return err
	}
	log.Info("Repeat: ", r.(*mopidy.BasicResponse).Result)
	return nil
}

func (m *MopidyPlayer) setRandom() error {
	if _, err := m.client.Call("core.tracklist.set_random", []bool{true}); err != nil {
		return err
	}
	r, err := m.client.Call("core.tracklist.get_random", nil)
	if err != nil {
		return err
	}
	log.Info("Random: ", r.(*mopidy.BasicResponse).Result)
	return nil
}
