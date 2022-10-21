package beater

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
	//"strconv"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/txtweet/test_velo_b1/config"
)

const (
	apiUrl  = "https://api.jcdecaux.com/"
	apiPath = "vls/v3/contracts"
	api_key = "76d6d73e00da651b6e90532a9ba2cfd1d2fabe72"

	selector = "velotest"
)

type ContractsRespnse struct {
	name     string `json:name`
	commName string `json:"commercial_name"`
}

// test_velo_b1 configuration.
type test_velo_b1 struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}
type apiResponsaData struct {
	Name       string   `json:"name"`
	Commercial string   `json:"commercial_name"`
	Cities     []string `json:"cities"`
}

// New creates an instance of test_velo_b1.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &test_velo_b1{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

// Run starts test_velo_b1.
func (bt *test_velo_b1) Run(b *beat.Beat) error {
	logp.Info("test_velo_b1 is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	counter := 1
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}
		var ParsedUrl *url.URL
		client := &http.Client{}

		ParsedUrl, err := url.Parse(apiUrl)
		if err != nil {
			logp.NewLogger(selector).Error("Unable to parse URL string")
			panic(err)
		}

		ParsedUrl.Path += apiPath

		parameters := url.Values{}
		parameters.Add("apiKey", api_key)

		ParsedUrl.RawQuery = parameters.Encode()

		logp.NewLogger(selector).Debug("Requesting Velov data: ", ParsedUrl.String())
		fmt.Println(ParsedUrl.String())
		req, err := http.NewRequest("GET", ParsedUrl.String(), nil)
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			logp.NewLogger(selector).Debug("Status code: ", res.StatusCode)
			logp.NewLogger(selector).Debug("Status code: ", res.Body)
			return fmt.Errorf("HTTP %v", res)
		}

		body, err := ioutil.ReadAll(res.Body)

		// check if the response is not an empty array
		if len(body) <= 2 {
			logp.NewLogger(selector).Debug("API call '", ParsedUrl.String(), "' returns 0 results. Response body: ", string(body))
			return nil
		}

		logp.NewLogger(selector).Debug(string(body))
		if err != nil {
			log.Fatal(err)
			return err
		}
		fmt.Println(string(body))

		var contractDatas []apiResponsaData
		err = json.Unmarshal(body, &contractDatas)
		if err != nil {
			//fmt.Println("error: %v", err)
			return err
		}

		logp.NewLogger(selector).Debug("Unmarshal-ed Owm data: ", contractDatas)

		for _, c := range contractDatas {
			event := beat.Event{
				Timestamp: time.Now(),
				Fields: common.MapStr{
					"type":     b.Info.Name,
					"contrats": c,
				},
			}
			bt.client.Publish(event)
			logp.Info("Event sent")
		}
		counter++
	}
}

// Stop stops test_velo_b1.
func (bt *test_velo_b1) Stop() {
	bt.client.Close()
	close(bt.done)
}
