package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
)

const stateFileName = "cdev-state.json"
const stateLockFileName = "cdev-state.lock"

func (b *Backend) LockState() error {
	stateLockFilePath := filepath.Join(b.Path, stateLockFileName)

	_, err := ioutil.ReadFile(stateLockFilePath)
	if err == nil {
		return fmt.Errorf("state is locked by another process")
	}
	err = ioutil.WriteFile(stateLockFilePath, []byte{}, os.ModePerm)
	return err
}

func (b *Backend) UnlockState() error {
	stateLockFilePath := filepath.Join(b.Path, stateLockFileName)
	log.Debugf("Unlocking local state. Path: '%v'", stateLockFilePath)
	return os.Remove(stateLockFilePath)
}

func (b *Backend) WriteState(stateData string) error {

	stateFilePath := filepath.Join(b.Path, stateFileName)
	log.Debugf("Updating local state. Project: '%v', path: '%v'", b.ProjectPtr.Name(), stateFilePath)

	err := ioutil.WriteFile(stateFilePath, []byte(stateData), os.ModePerm)
	return err
}

func (b *Backend) ReadState() (string, error) {
	stateFilePath := filepath.Join(b.Path, stateFileName)
	res, err := ioutil.ReadFile(stateFilePath)
	return string(res), err
}
