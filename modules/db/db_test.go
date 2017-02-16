package db

import (
	"os"
	"testing"
)

func cleanup(databasePath string, t *testing.T) {
	if err := remove(databasePath); err != nil {
		t.Errorf("failed to cleanup %s: %v", databasePath, err)
	}

	databaseLockPath := databasePath + ".lock"
	if err := remove(databaseLockPath); err != nil {
		t.Errorf("failed to cleanup %s: %v", databaseLockPath, err)
	}
}

func remove(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		err = os.Remove(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestDatabase(t *testing.T) {
	const databasePath = "test.db"
	cleanup(databasePath, t)
	defer cleanup(databasePath, t)

	var database Database
	var err error
	database, err = NewDatabase(databasePath)
	if err != nil {
		t.Errorf("failed to create database: %v", err)
		return
	}

	d1, err := database.GetOnlineUrDragon()
	if err != nil {
		t.Errorf("failed to get initial dragon: %v", err)
	}

	d2 := d1.NextGeneration()

	err = database.PutOnlineUrDragon(d2)
	if err != nil {
		t.Errorf("failed to save dragon data: %v", err)
	}

	d3, err := database.GetOnlineUrDragon()
	if err != nil {
		t.Errorf("failed to get dragon data: %v", err)
	}

	if d2.Generation != d3.Generation {
		t.Error("dragon data mismatch")
	}

	err = database.Close()
	if err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}
