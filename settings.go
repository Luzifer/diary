package main

import (
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type settings struct {
	DateFormat string `yaml:"DateFormat"`
	EditorCmd  string `yaml:"EditorCmd"`
	Encrypt    bool   `yaml:"Encrypt"`
}

func loadSettings(cmd *cobra.Command, args []string) error {
	data, err := ioutil.ReadFile(params.SettingsFile)
	if err != nil {
		return err
	}

	out := &settings{}
	if err := yaml.Unmarshal(data, out); err != nil {
		return err
	}

	set = out
	return nil
}
