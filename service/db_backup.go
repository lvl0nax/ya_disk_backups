package service

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type DbBackupInterface interface {
	MakeBackup() (string, error)
	RemoveBackupFile(filePath string)
}

type BackupService struct {
	dbName       string
	dbUserName   string
	dbDockerName string
	DbBackupInterface
}

func NewBackupService(dbName, dbUserName, dbDockerName string) *BackupService {
	return &BackupService{
		dbName:       dbName,
		dbUserName:   dbUserName,
		dbDockerName: dbDockerName,
	}
}

func (s *BackupService) MakeBackup() (string, error) {
	// File name for backup
	currentDate := time.Now().Format("20060102")
	fileName := fmt.Sprintf("%s_dump_%s.zip", s.dbName, currentDate)

	// build shell command
	dumpName := fmt.Sprintf("%s_dump_%s", s.dbName, currentDate)
	// docker exec -i postgres pg_dump --username db_username db_name
	dumpCommand := fmt.Sprintf("docker exec -t %s pg_dump -U %s %s > %s",
		s.dbDockerName, s.dbUserName, s.dbName, dumpName)
	zipCommand := fmt.Sprintf("zip %s %s", fileName, dumpName)
	removeCommand := fmt.Sprintf("rm %s", dumpName)

	mainCommand := strings.Join([]string{dumpCommand, zipCommand, removeCommand}, ";")
	cmd := exec.Command("sh", "-c", mainCommand)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	dumpPath := fmt.Sprintf("%s", fileName)
	if _, err := os.Stat(dumpPath); os.IsNotExist(err) {
		return "", errors.New(fmt.Sprintf("Dump %s not created", dumpPath))
	}
	return fileName, nil
}

func (s *BackupService) RemoveBackupFile(filePath string) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("rm %s;", filePath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Backup file remove err %s\n", err)
	}
}
