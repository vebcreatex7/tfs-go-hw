package config

import (
	"github.com/spf13/viper"
)

func Init() error {
	viper.AddConfigPath("/home/courage/tfs-go/tfs-go-hw/course_project/config")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
