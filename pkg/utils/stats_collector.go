package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/apex/log"
)

const StatsGatewayURL = "https://cdev-usage.cluster.dev/pushgateway"

type StatsExporter struct {
}

type CollectorRespond struct {
	Context struct {
		RequestID string `json:"request-id"`
	} `json:"context"`
}

func (e *StatsExporter) PushStats(stats interface{}) error {
	collectEnv := GetEnv("CDEV_COLLECT_USAGE_STATS", "true")
	if collectEnv == "false" {
		log.Debugf("Usage statistic sending is disabled. Skipping...")
		return nil
	}
	jsonBody, err := JSONEncode(map[string]interface{}{"stats": stats})
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest(http.MethodPost, StatsGatewayURL, bodyReader)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}

	client := http.Client{
		Timeout: 3 * time.Second,
	}
	log.Debugf("Usage stats:\n%v", string(jsonBody))
	res, err := client.Do(req)
	if err != nil {
		log.Warnf("Can't push usage stats: client: error making http request: %s\n", err)
		return err
	}
	respondBody, err := io.ReadAll(res.Body)
	respDecoded := &CollectorRespond{}
	JSONDecode(respondBody, respDecoded)
	if err != nil {
		return err
	}

	// log.Debugf("Usage statistics sent. Server respond:\n%v", string(respondBody))
	log.Debugf("Usage statistics request ID: %v", respDecoded.Context.RequestID)
	return nil
}
