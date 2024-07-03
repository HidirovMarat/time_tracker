package info

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const endpoint = "/info"

type UserInfoResponse struct {
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	Patronymic string `json:"patronymic"`
	Address    string `json:"address"`
}

type RInfo struct {
	UserInfo UserInfoResponse
}

func NewRI() *RInfo{ 
	return new(RInfo)
}

func (rInfo *RInfo) GetUserInfo(passportSerie int, passportNumber int, baseURL string) (*UserInfoResponse, error) {
	params := url.Values{}
	params.Add("passportSerie", strconv.Itoa(passportSerie))
	params.Add("passportNumber", strconv.Itoa(passportNumber))

	fullURL, err := url.Parse(baseURL)

	if err != nil {
		panic("not correct address = " + baseURL)
	}

	fullURL.Path += endpoint
	fullURL.RawQuery = params.Encode()

	resp, err := http.Get(fullURL.String())

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	var apiResp UserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	return &apiResp, nil
}
