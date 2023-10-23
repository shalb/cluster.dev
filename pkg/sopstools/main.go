package sopstools

import (
	"github.com/getsops/sops/v3/cmd/sops/formats"
	"github.com/getsops/sops/v3/decrypt"
)

func DecryptYaml(data []byte) ([]byte, error) {
	formatFmt := formats.FormatForPathOrString("", "yaml")
	return decrypt.DataWithFormat(data, formatFmt)
}

func ReadAndDecrypt(filename string) ([]byte, error) {
	return decrypt.File(filename, "yaml")
}
