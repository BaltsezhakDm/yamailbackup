package backup

import (
	"log"
	"testing"

	"github.com/KrasovD/yamailbackup/internal/utils"
)

func TestBackup(t *testing.T) {
	config, err := utils.LoadConfig("../../config/config.yaml")
	if err != nil {
		log.Fatal("Error loading config:", err)
	}
	err = GetListCloud(config)
	if err != nil {
		log.Fatal("Error connecting to cloud:", err)
	}

}
