package transmission

import (
  "net/http"
  "fmt"
  "io/ioutil"
  "encoding/json"
)

type Client struct {
  Host string
  Port int32
  token string
}

func NewClient(host string, port int32) *Client {
  return &Client{ host, port, "" }
}

func (client *Client) refresh() error {
  req, err := RefreshRequest(Connection{client.Host, client.Port}).ToRequest()
  if err != nil {
    panic(err)
  }

  httpClient := &http.Client{}
  response, err := httpClient.Do(req)

  if err != nil {
    return err
  }

  token := response.Header["X-Transmission-Session-Id"][0]
  if token == "" {
    return fmt.Errorf("Failed to authenticate: couldn't receive session ID token.")
  }

  client.token = token
  return nil
}

func (client *Client) perform(builder func(Connection, string)(*http.Request, error)) ([]byte, error) {
  var err error
  if client.token == "" {
    err = client.refresh()
  }

  if err != nil {
    return nil, err
  }

  req, err := builder(Connection{client.Host, client.Port}, client.token)

  if err != nil {
    return nil, err
  }

  httpClient := &http.Client{}
  response, err := httpClient.Do(req)

  if err != nil {
    return nil, err
  }


  if header, present := response.Header["X-Transmission-Session-Id"]; present {
    client.token = header[0]
  } else {
    err = fmt.Errorf("No token present")
  }

  if (response.StatusCode != 200) {
    return nil, fmt.Errorf("Bad response code: %d", response.StatusCode)
  }

  defer response.Body.Close()
  body, err := ioutil.ReadAll(response.Body)

  return body, err
}

func (client *Client) performJson(
  builder func(Connection, string)(*http.Request, error),
  response TResponse) error {
  body, err := client.perform(builder)

  if err != nil {
    return err
  }

  jsonErr := json.Unmarshal(body, &response)
  if jsonErr != nil {
    return fmt.Errorf("Error: %s", jsonErr)
  }

  if (response.Result() != "success") {
    return fmt.Errorf("Error: %s", response.Result())
  }

  return nil
}

func (client *Client) performWithoutData(builder func(Connection, string)(*http.Request, error)) error {
  var response GenericResponse
  err := client.performJson(builder, &response)

  if err != nil {
    return err
  }

  if (response.Result() != "success") {
    return fmt.Errorf("Error: %s", response.Result())
  }

  return nil
}
