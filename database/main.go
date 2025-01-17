package database

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/safing/portbase/utils"
	"github.com/tevino/abool"
)

const (
	databasesSubDir = "databases"
)

var (
	initialized = abool.NewBool(false)

	shuttingDown   = abool.NewBool(false)
	shutdownSignal = make(chan struct{})

	rootStructure      *utils.DirStructure
	databasesStructure *utils.DirStructure
)

// Initialize initializes the database at the specified location. Supply either a path or dir structure.
func Initialize(dirPath string, dirStructureRoot *utils.DirStructure) error {
	if initialized.SetToIf(false, true) {

		if dirStructureRoot != nil {
			rootStructure = dirStructureRoot
		} else {
			rootStructure = utils.NewDirStructure(dirPath, 0755)
		}

		// ensure root and databases dirs
		databasesStructure = rootStructure.ChildDir(databasesSubDir, 0700)
		err := databasesStructure.Ensure()
		if err != nil {
			return fmt.Errorf("could not create/open database directory (%s): %s", rootStructure.Path, err)
		}

		err = loadRegistry()
		if err != nil {
			return fmt.Errorf("could not load database registry (%s): %s", filepath.Join(rootStructure.Path, registryFileName), err)
		}

		// start registry writer
		go registryWriter()

		return nil
	}
	return errors.New("database already initialized")
}

// Shutdown shuts down the whole database system.
func Shutdown() (err error) {
	if shuttingDown.SetToIf(false, true) {
		close(shutdownSignal)
	} else {
		return
	}

	controllersLock.RLock()
	defer controllersLock.RUnlock()

	for _, c := range controllers {
		err = c.Shutdown()
		if err != nil {
			return
		}
	}
	return
}

// getLocation returns the storage location for the given name and type.
func getLocation(name, storageType string) (string, error) {
	location := databasesStructure.ChildDir(name, 0700).ChildDir(storageType, 0700)
	// check location
	err := location.Ensure()
	if err != nil {
		return "", fmt.Errorf(`failed to create/check database dir "%s": %s`, location.Path, err)
	}
	return location.Path, nil
}
