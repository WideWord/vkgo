package vk


import(
	"net/http"
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

	query := fmt.Sprintf("https://oauth.vk.com/access_token?client_id=%s&client_secret=%s&v=5.24&grant_type=client_credentials", client.appId, client.appSecret)

	log.Printf("%s\n", query)

	resp, err := http.Get(query)
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

func (client *Client) PlainCall(method string, params string, response interface{}) (err error) {
	query := fmt.Sprintf("https://api.vk.com/method/%s?v=5.24&%s", method, params)

	log.Printf("%s\n", query)

	resp, err := http.Get(query)
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

func (client *Client) SequreCall(method string, params string, response interface{}) (err error) {
	return client.Call(method, fmt.Sprintf("access_token=%s&%s", client.serverAccessToken, params), response)
}

func (client *Client) Call(method string, params string, response interface{}) (err error) {
	return client.PlainCall(method, params, response)
}
