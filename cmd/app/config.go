package app

import (
	"encoding/json"
	"errors"
	"time"
)

const ConfigLocation = "/etc/player-resolver/config.json"

type Config struct {
	MojangAPIQueryInterval Duration `json:"mojangApiQueryInterval"`
	Host                   string   `json:"host"`
	Port                   int      `json:"port"`
	CassandraHost          string   `json:"cassandraHost"`
}

type Duration struct {
	time.Duration
}

func NewDefaultConfig() *Config {
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}
