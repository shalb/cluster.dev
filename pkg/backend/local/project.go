package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
)

func (b *Backend) LockState() error {
	stateLockFilePath := filepath.Join(b.Path, "state.lock")

	_, err := ioutil.ReadFile(stateLockFilePath)
	if err == nil {
		return fmt.Errorf("state is locked by another process")
	}
	err = ioutil.WriteFile(stateLockFilePath, []byte{}, os.ModePerm)
	return err
}

func (b *Backend) UnlockState() error {
	stateLockFilePath := filepath.Join(b.Path, "state.lock")
	log.Debugf("Unlocking local state. Path: '%v'", stateLockFilePath)
	return os.Remove(stateLockFilePath)
}

func (b *Backend) WriteState(stateData string) error {

	stateFilePath := filepath.Join(b.Path, "state.json")
	log.Debugf("Updating local state. Project: '%v', path: '%v'", b.ProjectPtr.Name(), stateFilePath)

	err := ioutil.WriteFile(stateFilePath, []byte(stateData), os.ModePerm)
	return err
}

func (b *Backend) ReadState() (string, error) {
	stateFilePath := filepath.Join(b.Path, "state.json")
	log.Debugf("Reading local state. Project: '%v', bucket: '%v'", b.ProjectPtr.Name(), stateFilePath)
	res, err := ioutil.ReadFile(stateFilePath)
	return string(res), err
}
