package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/spf13/viper"

	"github.com/lvl0nax/yadisk_db_backuper/service"
)

const (
	newFolder = "backups"
)

type Config struct {
	DBName         string `mapstructure:"DB_NAME"`
	DBUsername     string `mapstructure:"DB_USERNAME"`
	DBDockerName   string `mapstructure:"DB_DOCKERNAME"`
	AuthToken      string `mapstructure:"AUTH_TOKEN"`
	YaAppName      string `mapstructure:"YA_APP_NAME"`
	KeepBackupsNum int    `mapstructure:"BACKUPS_NUM"`
}

func initConfig() (config Config, err error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)

	v := reflect.ValueOf(config)
	typeOfS := v.Type()
	fmt.Println("========================CONFIG=========================")
	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("%s = \t%v\n", typeOfS.Field(i).Name, v.Field(i).Interface())
	}
	fmt.Println("=======================================================")

	return
}

func main() {
	config, err := initConfig()
	if err != nil {
		log.Fatalf("Config initializing error %s", err)
		return
	}
	fmt.Println("=> Config Initialized")

	yaService := service.NewYaService(config.AuthToken, config.YaAppName)
	err = yaService.CreateFolder(newFolder)
	if err != nil {
		log.Fatalf("Folder creation failed %s", err)
		return
	}
	fmt.Println("=> 'backups' folder created")

	backupService := service.NewBackupService(config.DBName, config.DBUsername, config.DBDockerName)
	fileName, err := backupService.MakeBackup()
	if err != nil {
		log.Fatalf("Dump file creation error %s", err)
		return
	}
	fmt.Printf("=> Backup file %s created\n", fileName)

	err = yaService.UploadFile(fileName, newFolder+"/"+fileName)
	if err != nil {
		log.Fatalf("File upload failed %s", err)
		return
	}
	fmt.Printf("=> Backup file %s uploaded to Yandex Disk\n", fileName)

	err = yaService.RemoveOldBackups(newFolder, config.KeepBackupsNum)
	if err != nil {
		log.Fatalf("Old backups clean up failed %s", err)
		return
	}
	fmt.Println("=> Old backups cleaned up")

	backupService.RemoveBackupFile(fileName)
	fmt.Println("=> Recent backup file removed from the machine")
}
