package vk


import(
	"net/http"
	"net/url"
	"fmt"
	"errors"
	"io/ioutil"
	"encoding/json"
	"log"
)


type Client struct {
	appId string
	appSecret string
	serverAccessToken string
}

func NewClient() *Client {
	client := new(Client)
	return client
}

func (client *Client) AuthServer(id string, secret string) (err error) {
	client.appId = id
	client.appSecret = secret

	query, err := url.Parse("https://oauth.vk.com/access_token")
	params := url.Values{}
	params.Set("client_id", client.appId)
	params.Set("client_secret", client.appSecret)
	params.Set("v", "5.24")
	params.Set("grant_type", "client_credentials")
	query.RawQuery = params.Encode()

	url := query.String()

	log.Printf("%s\n", url)

	resp, err := http.Get(url)
	if err != nil { return }

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil { return }

	type ParsedData struct {
		Error string
		Access_token string
	}
	var parsedData ParsedData

	err = json.Unmarshal(data, &parsedData)
	if err != nil { return }

	if parsedData.Error != "" { return errors.New(parsedData.Error) }

	client.serverAccessToken = parsedData.Access_token

	return nil
}

func (client *Client) PlainCall(method string, params url.Values, response interface{}) (err error) {

	query, err := url.Parse("https://api.vk.com/")

	query.Scheme = "https"
	query.Host = "api.vk.com"
	query.Path = fmt.Sprintf("/method/%s", method)
	query.RawQuery = params.Encode()

	url := query.String()

	log.Printf("%s\n", url)

	resp, err := http.Get(url)
	if err != nil { return }

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil { return }

	type ParsedData struct {
		Error string
		Response interface{}
	}
	var parsedData ParsedData

	parsedData.Response = response

	err = json.Unmarshal(data, &parsedData)
	if err != nil { return }

	if parsedData.Error != "" { return errors.New(parsedData.Error) }

	return nil
}

func (client *Client) SecureCall(method string, params url.Values, response interface{}) (err error) {
	params.Add("access_token", client.serverAccessToken)
	return client.Call(method, params, response)
}

func (client *Client) Call(method string, params url.Values, response interface{}) (err error) {
	return client.PlainCall(method, params, response)
}
