package project

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func createMarker(t string) string {
	const markerLen = 10
	hash := randSeq(markerLen)
	return fmt.Sprintf("%s.%s.%s", hash, t, hash)
}

func printVersion() string {
	return "master"
}

func removeDirContent(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func findModule(infra, name string, modsList map[string]*Module) *Module {
	mod, exists := modsList[fmt.Sprintf("%s.%s", infra, name)]
	// log.Printf("Check Mod: %s, exists: %v, list %v", name, exists, modsList)
	if !exists {
		return nil
	}
	return mod
}
