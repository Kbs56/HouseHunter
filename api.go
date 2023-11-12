package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type ResStruct struct {
	Data struct {
		HomeSearch struct {
			Results []struct {
				Location struct {
					Address struct {
						City       string `json:"city"`
						Line       string `json:"line"`
						PostalCode string `json:"postal_code"`
						StateCode  string `json:"state_code"`
					} `json:"address"`
				} `json:"location"`
				Description struct {
					Sqft  int `json:"sqft"`
					Beds  int `json:"beds"`
					Baths int `json:"baths"`
				} `json:"description"`
				Href               string `json:"href"`
				ListPrice          int    `json:"list_price"`
				PriceReducedAmount int    `json:"price_reduced_amount"`
				LastSoldPrice      int    `json:"last_sold_price"`
				ListDate           string `json:"list_date"`
				Status             string `json:"status"`
			} `json:"results"`
		} `json:"home_search"`
	} `json:"data"`
}

type ReqBody struct {
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	PostalCode string   `json:"postal_code"`
	Status     []string `json:"status"`
	SortFields struct {
		Direction string `json:"direction"`
		Field     string `json:"field"`
	} `json:"sort_fields"`
	ListPrice struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"list_price"`
	Beds struct {
		Min int `json:"min"`
	} `json:"beds"`
	Baths struct {
		Min int `json:"min"`
	} `json:"baths"`
	Sqft struct {
		Min int `json:"min"`
	} `json:"sqft"`
}

func callApi(
	searchArea string,
	priceMin int,
	priceMax int,
	bedMin int,
	bathMin int,
	sqftMin int,
	status []string,
	numResults int,
	ch chan<- string,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	url := "https://realtor.p.rapidapi.com/properties/v3/list"

	reqBody := &ReqBody{
		Limit:      numResults,
		Offset:     0,
		PostalCode: searchArea,
		Status:     status,
		SortFields: struct {
			Direction string "json:\"direction\""
			Field     string "json:\"field\""
		}{Direction: "desc", Field: "list_date"},
		ListPrice: struct {
			Min int "json:\"min\""
			Max int "json:\"max\""
		}{Min: priceMin, Max: priceMax},
		Beds: struct {
			Min int "json:\"min\""
		}{Min: bedMin},
		Baths: struct {
			Min int "json:\"min\""
		}{Min: bathMin},
		Sqft: struct {
			Min int "json:\"min\""
		}{Min: sqftMin},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))

	req.Header.Add("content-type", "application/json")
	req.Header.Add("X-RapidAPI-Key", os.Getenv("realtorApiKey"))
	req.Header.Add("X-RapidAPI-Host", "realtor.p.rapidapi.com")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	resp, _ := io.ReadAll(res.Body)

	err = readJson(resp, ch)
	if err != nil {
		panic(err)
	}
}

func readJson(body []byte, ch chan<- string) error {
	res := &ResStruct{}
	json.Unmarshal(body, res)
	for i := 0; i < len(res.Data.HomeSearch.Results); i++ {
		fullStreet := res.Data.HomeSearch.Results[i].Location.Address.Line
		city := res.Data.HomeSearch.Results[i].Location.Address.City
		state := res.Data.HomeSearch.Results[i].Location.Address.StateCode
		zip := res.Data.HomeSearch.Results[i].Location.Address.PostalCode
		listDate := res.Data.HomeSearch.Results[i].ListDate
		formattedDate, err := formatDate(listDate)
		if err != nil {
			panic(err)
		}
		ch <- fmt.Sprintf("%s, %s, %s, %s\nLink: %s\nList Price: %d\nList Date: %s\nStatus: %s\nPrice Reduced Amount: %d\nLast Sold Price: %d\nSqft: %d\nBeds: %d\nBaths: %d\n",
			fullStreet,
			city,
			state,
			zip,
			res.Data.HomeSearch.Results[i].Href,
			res.Data.HomeSearch.Results[i].ListPrice,
			formattedDate, res.Data.HomeSearch.Results[i].Status,
			res.Data.HomeSearch.Results[i].PriceReducedAmount,
			res.Data.HomeSearch.Results[i].LastSoldPrice,
			res.Data.HomeSearch.Results[i].Description.Sqft,
			res.Data.HomeSearch.Results[i].Description.Beds,
			res.Data.HomeSearch.Results[i].Description.Baths)
	}
	return nil
}

func formatDate(dateString string) (string, error) {
	dateObj, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		return "", err
	}
	formattedDate := dateObj.Format("06-01-02")
	return formattedDate, nil
}
