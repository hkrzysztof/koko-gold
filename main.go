// koko shows the current gold prices in several currencies and units
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
)

// checks for errors
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// getPrice grabs current gold price from API and returns a body of the response
func getPrice() []byte {
	resp, err := http.Get("http://api.nbp.pl/api/cenyzlota")
	check(err)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	check(err)

	return body
}

// getPrice grabs current currency rate from API and returns a body of the response
func getRate(cur string) []byte {
	url := "http://api.nbp.pl/api/exchangerates/rates/a/" + cur
	resp, err := http.Get(url)
	check(err)

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		check(err)
		return body
	} else {
		log.Fatal(errors.New("wrong currency"))
		return []byte{}
	}
}

// parseJsonPrice parses raw json price response into a float64
func parseJsonPrice(rawJs []byte) float64 {
	type ApiPrice []struct {
		Date  string  `json:"data"`
		Price float64 `json:"cena"`
	}
	resp := ApiPrice{}

	err := json.Unmarshal(rawJs, &resp)
	check(err)

	return resp[0].Price
}

// parseJsonRate parses raw json currency rate response into a float64
func parseJsonRate(rawJs []byte) float64 {
	type ApiRate struct {
		Table    string `json:"table"`
		Currency string `json:"currency"`
		Code     string `json:"code"`
		Rates    []struct {
			No            string  `json:"no"`
			EffectiveDate string  `json:"effectiveDate"`
			Mid           float64 `json:"mid"`
		} `json:"rates"`
	}
	resp := ApiRate{}

	err := json.Unmarshal(rawJs, &resp)
	check(err)

	return resp.Rates[0].Mid
}

// converseUnit converses unit according to a flag
func converseUnit(value float64, unit string) float64 {
	switch unit {
	case "kg":
		return value * 1000
	case "g":
		return value
	case "lbs":
		return value * 453.59
	case "", "oz":
		return value * 28.35
	default:
		log.Fatal(errors.New("wrong unit"))
		return 0.0
	}
}

// converseUnit converses currency according to a flag
func converseCurrency(value float64, cur string) float64 {
	if cur == "" {
		return value / parseJsonRate(getRate("usd"))
	} else {
		return value / parseJsonRate(getRate(cur))
	}
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "gold",
				Usage: "shows current price of gold",
				Action: func(c *cli.Context) error {
					parsedPrice := parseJsonPrice(getPrice())
					unit := c.String("unit")
					cur := c.String("cur")
					price := converseCurrency(converseUnit(parsedPrice, unit), cur)
					fmt.Printf("Current gold price: %0.2f %s per %s", price, cur, unit)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "unit",
						Usage: "choose unit by short`name`",
						Value: "oz",
					},
					&cli.StringFlag{
						Name:  "cur",
						Usage: "choose currency by short`name`",
						Value: "usd",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	check(err)
}
