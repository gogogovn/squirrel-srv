package acckit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func getAccessToken(fbAppID string, fbSecretKey string, authCode string) (*AccessTokenResponse, error) {
	kitClient := http.Client{
		Timeout: time.Second * 5, // Maximum 5 seconds
	}

	graphURL := fmt.Sprintf(accessTokenUrl, authCode, fbAppID, fbAppSecret)

	req, err := http.NewRequest(http.MethodGet, graphURL, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := kitClient.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		errResponse := struct {
			E Error `json:"error"`
		}{}
		if err := json.Unmarshal(body, &errResponse); err != nil {
			return nil, err
		}
		return nil, errors.New(errResponse.E.Message)

	}

	response := AccessTokenResponse{}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func getMe(accessToken string, secretKey string) (*GraphAPIResponse, error) {
	kitClient := http.Client{
		Timeout: time.Second * 5, // Maximum 5 seconds
	}

	secretProof := hmac256(accessToken, secretKey)

	graphURL := fmt.Sprintf(getMeUrl, accessToken, secretProof)

	req, err := http.NewRequest(http.MethodGet, graphURL, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := kitClient.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		errResponse := struct {
			E Error `json:"error"`
		}{}
		if err := json.Unmarshal(body, &errResponse); err != nil {
			return nil, err
		}
		return nil, errors.New(errResponse.E.Message)

	}

	response := GraphAPIResponse{}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
