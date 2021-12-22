package utils

import (
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/apparentlymart/go-cidr/cidr"
)

func CidrSubnet(netCidr string, newbits int, netnum interface{}) (ret string, err error) {
	var netnumInt int
	switch v := netnum.(type) {
	case int:
		netnumInt = v
	case int64:
		netnumInt = int(v)
	case int16:
		netnumInt = int(v)
	case int32:
		netnumInt = int(v)
	default:
		return "", fmt.Errorf("unsupported type %T", netnum)
	}

	_, network, err := net.ParseCIDR(netCidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR expression: %s", err)
	}

	newNetwork, err := cidr.SubnetBig(network, newbits, big.NewInt(int64(netnumInt)))
	if err != nil {
		return "", err
	}

	return newNetwork.String(), nil
}

func GetFileByUrl(srcUrl string) (string, error) {

	_, err := url.ParseRequestURI(srcUrl)
	if err != nil {
		return "", err
	}
	// Get the data
	resp, err := http.Get(srcUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
