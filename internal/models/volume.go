package models

import "time"

type Volume struct {
	Volumes []struct {
		CreatedAt  time.Time         `json:"CreatedAt"`
		Name       string            `json:"Name"`
		Driver     string            `json:"Driver"`
		Mountpoint string            `json:"Mountpoint"`
		Labels     map[string]string `json:"Labels"`
		Scope      string            `json:"Scope"`
		Options    struct {
			Device string `json:"device"`
			O      string `json:"o"`
			Type   string `json:"type"`
		} `json:"Options"`
	} `json:"Volumes"`
	Warnings []interface{} `json:"Warnings"`
}
