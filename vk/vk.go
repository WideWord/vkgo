package vk


import(
	"net/http"
	"net/url"
	"fmt"
	"errors"
	"io/ioutil"
	"encoding/json"
	"log"
	"crypto/md5"
	"encoding/hex"
)


type Client struct {
	appId string
	appSecret string
	serverAccessToken string
	LogEverything bool
}

func NewClient(appId string, appSecret string) *Client {
	client := new(Client)
	client.appId = appId
	client.appSecret = appSecret
	client.LogEverything = false
	return client
}

func (client *Client) authServer(hc *http.Client) (err error) {

	defer func() {
		if r := recover(); r != nil {
			if err == nil {
				log.Printf("[VK] nil err")
			} else if client.LogEverything {
				log.Printf("[VK] authServer error %s", err.Error())
			}
		}
	}()

	query, err := url.Parse("https://oauth.vk.com/access_token")
	params := url.Values{}
	params.Set("client_id", client.appId)
	params.Set("client_secret", client.appSecret)
	params.Set("v", "5.24")
	params.Set("grant_type", "client_credentials")
	query.RawQuery = params.Encode()

	url := query.String()

	if client.LogEverything {
		log.Printf("[VK] GET %s", url)
	}

	resp, err := hc.Get(url)
	if err != nil { panic(err) }

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil { panic(err) }

	var parsedData struct {
		Error string
		Access_token string
	}


	err = json.Unmarshal(data, &parsedData)
	if err != nil {
		err = errors.New(fmt.Sprintf("%s\n====\n%s\n====", err.Error(), data))
		panic(err)
	}

	if parsedData.Error != "" {
		err = errors.New(parsedData.Error)
		panic(err)
	}

	client.serverAccessToken = parsedData.Access_token

	return nil
}

func (client *Client) PlainCall(hc *http.Client, method string, params url.Values, response interface{}) (err error) {

	defer func() {
		if r := recover(); r != nil {
			if client.LogEverything {
				log.Printf("[VK] error %s", err.Error())
			}
		}
	}()

	query, err := url.Parse("https://api.vk.com/")

	query.Scheme = "https"
	query.Host = "api.vk.com"
	query.Path = fmt.Sprintf("/method/%s", method)
	query.RawQuery = params.Encode()

	url := query.String()

	if client.LogEverything {
		log.Printf("[VK] GET %s", url)
	}

	resp, err := hc.Get(url)
	if err != nil { panic(err) }

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil { panic(err) }

	var parsedData struct {
		Error struct {
			Error_code int
			Error_msg string
		}
		Response interface{}
	}

	parsedData.Response = response

	err = json.Unmarshal(data, &parsedData)
	if err != nil {
		err = errors.New(fmt.Sprintf("%s\n====\n%s\n====", err.Error(), data))
		panic(err)
	}

	if parsedData.Error.Error_msg != "" { err = errors.New(parsedData.Error.Error_msg); panic(err) }

	return nil
}

func (client *Client) AuthCall(hc *http.Client, method string, params url.Values, response interface{}) (err error) {
	if client.serverAccessToken == "" {
		client.authServer(hc)
	}
	params.Add("access_token", client.serverAccessToken)
	return client.Call(hc, method, params, response)
}

func (client *Client) SecureCall(hc *http.Client, method string, params url.Values, response interface{}) (err error) {
	params.Add("client_secret", client.appSecret)
	return client.AuthCall(hc, method, params, response)
}

func (client *Client) Call(hc *http.Client, method string, params url.Values, response interface{}) (err error) {
	return client.PlainCall(hc, method, params, response)
}

func (client *Client) CheckUserAuthKey(user int, key string) bool {
	str := fmt.Sprintf("%s_%d_%s", client.appId, user, client.appSecret)
	hash := md5.Sum([]byte(str))
	return key == hex.EncodeToString(hash[:])
}

