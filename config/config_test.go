package config

import (
	"os"
	"testing"
)

func TestLoadConfigFromFile(t *testing.T) {
	os.Setenv(MappingFile, "../tests/rabbit_to_sns.json")

	pairs, err := loadPairsFromFile()
	if err != nil {
		t.Errorf("error loading config from file: %s", err.Error())
	}
	if len(pairs) < 1 {
		t.Errorf("empty config loaded from file: %s", err.Error())
	}
}

func TestLoadConfigFromVault(t *testing.T) {
	pairs, err := loadPairsFromVault()
	if err != nil {
		t.Errorf("error loading config from Vault: %s", err.Error())
	}
	if len(pairs) < 1 {
		t.Errorf("empty config loaded from Vault: %s", err.Error())
	}
}
