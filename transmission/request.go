package transmission

import (
  "net/http"
  "encoding/json"
  "bytes"
  "fmt"
)

type Requester interface {
  ToRequest() (*http.Request, error)
}

type Connection struct {
  Host string
  Port int32
}

type TRequest struct {
  Connection Connection
  Method string
  Token string
  Arguments interface{}
}

func (request TRequest) ToRequest() (*http.Request, error) {
  var body = make(map[string]interface{})
  if request.Method != "" {
    body["method"] = request.Method
  }
  if request.Arguments != nil {
    body["arguments"] = request.Arguments
  }

  byteData, err := json.Marshal(body)
  if err != nil {
    return nil, err
  }
  reader := bytes.NewBuffer(byteData)

  req, err := http.NewRequest(
    "POST",
    fmt.Sprintf("http://%s:%d/transmission/rpc/", request.Connection.Host, request.Connection.Port),
    reader)

  if err != nil {
    return nil, err
  }

  if request.Token != "" {
    req.Header.Add("X-Transmission-Session-Id", request.Token)
  }

  return req, nil
}

type RequestBuilder func(conn Connection, token string)(*http.Request, error)
