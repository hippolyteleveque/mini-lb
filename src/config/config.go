package config

import (
    "bufio"
    "gopkg.in/yaml.v3"
    "os"
)

type Config struct {
	Port int `yaml:"Port"`
	Servers []string `yaml:"Servers"`
}

func Parse() (*Config, error) {
	file, err := os.Open("config.yaml")
	if err != nil {
		return nil, err
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	data := make([]byte, stat.Size())
	reader := bufio.NewReader(file)

	_, err = reader.Read(data)
	if err != nil {
		return nil, err
	}

	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}