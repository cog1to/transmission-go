package transmission

import (
  "net/http"
  "encoding/json"
  "bytes"
  "fmt"
  "io/ioutil"
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

/* Data */

const (
  TR_ETA_NOT_AVAIL = -1
  TR_ETA_UNKNOWN = -2
)

const (
  TR_STATUS_STOPPED = 0 /* Torrent is stopped */
  TR_STATUS_CHECK_WAIT = 1 /* Queued to check files */
  TR_STATUS_CHECK = 2 /* Checking files */
  TR_STATUS_DOWNLOAD_WAIT = 3 /* Queued to download */
  TR_STATUS_DOWNLOAD = 4 /* Downloading */
  TR_STATUS_SEED_WAIT = 5 /* Queued to seed */
  TR_STATUS_SEED = 6 /* Seeding */
)

type TorrentListItem struct {
  Id int64              `json:"id"`
  Name string           `json:"name"`
  UploadSpeed float32   `json:"rateUpload"`
  DownloadSpeed float32 `json:"rateDownload"`
  Ratio float32         `json:"uploadRatio"`
  Eta int32             `json:"eta"`
  SizeWhenDone int64    `json:"sizeWhenDone"`
  LeftUntilDone int64   `json:"leftUntilDone"`
  Status int8           `json:"status"`
}

type TorrentListResponseArguments struct {
  Torrents []TorrentListItem `json:"torrents"`
}

type TorrentListResponse struct {
  Result string                          `json:"result"`
  Tag string                             `json:"tag"`
  Arguments TorrentListResponseArguments `json:"arguments"`
}

type TorrentErrorResponse struct {
  Result string                    `json:"result"`
}

/* Requests */

func RefreshRequest(conn Connection) TRequest {
  return TRequest{
    conn,
    "",
    "",
    nil}
}

func ListRequest(conn Connection, token string) TRequest {
  return TRequest{
    conn,
    "torrent-get",
    token,
    map[string]interface{} { "fields": []string{
      "error",
      "errorString",
      "eta",
      "id",
      "leftUntilDone",
      "name",
      "rateDownload",
      "rateUpload",
      "sizeWhenDone",
      "status",
      "uploadRatio"}}}
}

func DeleteRequest(conn Connection, token string, ids []int64, withData bool) TRequest {
  return TRequest{
    conn,
    "torrent-remove",
    token,
    map[string]interface{} { "ids": ids, "delete-local-data": withData }}
}

func AddRequest(conn Connection, token string, filename string, downloadDir string, paused bool) TRequest {
  return TRequest{
    conn,
    "torrent-add",
    token,
    map[string]interface{} { "filename": filename, "download-dir": downloadDir, "paused": paused }}
}
/* Client */

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

func (client *Client) perform(builder func(Connection, string)(*http.Request, error)) (int, []byte, error) {
  var err error
  if client.token == "" {
    err = client.refresh()
  }

  if err != nil {
    return 200, nil, err
  }

  req, err := builder(Connection{client.Host, client.Port}, client.token)

  if err != nil {
    return 200, nil, err
  }

  httpClient := &http.Client{}
  response, err := httpClient.Do(req)

  if err != nil {
    return 200, nil, err
  }


  if header, present := response.Header["X-Transmission-Session-Id"]; present {
    client.token = header[0]
  } else {
    err = fmt.Errorf("No token present")
  }

  defer response.Body.Close()
  body, err := ioutil.ReadAll(response.Body)

  return response.StatusCode, body, err
}

func (client *Client) List() (*[]TorrentListItem, error) {
  code, body, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return ListRequest(conn, token).ToRequest()
  })

  if err != nil {
    return nil, err
  }

  if (code != 200) {
    return nil, fmt.Errorf("Bad response code: %d", code)
  }

  var listResponse TorrentListResponse
  json.Unmarshal(body, &listResponse)

  return &listResponse.Arguments.Torrents, err
}

func (client *Client) Delete(ids []int64, withData bool) error {
  code, _, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return DeleteRequest(conn, token, ids, withData).ToRequest()
  })

  if code != 200 {
    return fmt.Errorf("Bad response code: %d", code)
  }

  return err
}

func (client *Client) AddTorrent(url string, path string) (error) {
  code, body, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return AddRequest(conn, token, url, path, false).ToRequest()
  })

  if code != 200 {
    return fmt.Errorf("Bad response code: %d", code)
  }

  if err != nil {
    return err
  }

  var errorResponse TorrentErrorResponse
  jsonErr := json.Unmarshal(body, &errorResponse)
  if jsonErr == nil {
    return fmt.Errorf("%s", errorResponse.Result)
  }

  return nil
}
