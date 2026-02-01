package internal

import (
	"encoding/json"
	"errors"
	"os"
)

const StateFile = "state.json"

func LoadState() (*State, error) {
	file, err := os.ReadFile(StateFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &State{}, nil // first run
		}
		return nil, err // actual read error
	}

	var state State
	if err := json.Unmarshal(file, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func SaveState(state *State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(StateFile, data, 0644)
}
