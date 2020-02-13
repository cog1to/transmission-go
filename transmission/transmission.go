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

const (
  TR_PRIORITY_NORMAL = 0
  TR_PRIORITY_HIGH = 1
  TR_PRIORITY_LOW = -1
)

/* Responses */

type TorrentListItem struct {
  TorrentId int         `json:"id"`
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

type TorrentAddedInfo struct {
  HashString string `json:"hashString"`
  Id int            `json:"id"`
  Name string       `json:"name"`
}

type TorrentAddResponseArguments struct {
  Torrent *TorrentAddedInfo `json:"torrent-added"`
}

type TorrentAddResponse struct {
  Result string                         `json:"result"`
  Arguments TorrentAddResponseArguments `json:"arguments"`
}

type TorrentFileInternal struct {
  BytesCompleted int64 `json:"bytesCompleted"`
  Length int64         `json:"length"`
  Name string          `json:"name"`
}

type TorrentFileStatsInternal struct {
  BytesCompleted int64 `json:"bytesCompleted"`
  Wanted bool          `json:"wanted"`
  Priority int         `json:"priority"`
}

type TorrentFile struct {
  Number int
  BytesCompleted int64
  Length int64
  Name string
  Wanted bool
  Priority int
}

type TorrentDetailsInternal struct {
  Id int                                `json:"id"`
  Name string                           `json:"name"`
  UploadSpeed float32                   `json:"rateUpload"`
  DownloadSpeed float32                 `json:"rateDownload"`
  Ratio float32                         `json:"uploadRatio"`
  Eta int32                             `json:"eta"`
  SizeWhenDone int64                    `json:"sizeWhenDone"`
  LeftUntilDone int64                   `json:"leftUntilDone"`
  Status int8                           `json:"status"`
  DownloadLimit int                     `json:"downloadLimit"`
  DownloadLimited bool                  `json:"downloadLimited"`
  UploadLimit int                       `json:"uploadLimit"`
  UploadLimited bool                    `json:"uploadLimited"`
  Files *[]TorrentFileInternal          `json:"files"`
  FileStats *[]TorrentFileStatsInternal `json:"fileStats"`
}

type TorrentDetails struct {
  Id int
  Name string
  UploadSpeed float32
  DownloadSpeed float32
  Ratio float32
  Eta int32
  SizeWhenDone int64
  LeftUntilDone int64
  Status int8
  DownloadLimit int
  DownloadLimited bool
  UploadLimit int
  UploadLimited bool
  Files []TorrentFile
}

type TorrentDetailsResponseArguments struct {
  Torrents *[]TorrentDetailsInternal `json:"torrents"`
}

type TorrentDetailsResponse struct {
  Result string                             `json:"result"`
  Tag string                                `json:"tag"`
  Arguments TorrentDetailsResponseArguments `json:"arguments"`
}

type SessionSettings struct {
  UploadSpeedLimit int           `json:"speed-limit-up"`
  UploadSpeedLimitEnabled bool   `json:"speed-limit-up-enabled"`
  DownloadSpeedLimit int         `json:"speed-limit-down"`
  DownloadSpeedLimitEnabled bool `json:"speed-limit-down-enabled"`
}

type SessionSettingsResponse struct {
  Result string              `json:"result"`
  Tag string                 `json:"tag"`
  Arguments *SessionSettings `json:"arguments"`
}

type SetSessionSettingsResponse struct {
  Result string `json:"result"`
  Tag string    `json:"tag"`
}

/* Helpers */

func (file TorrentFile) Id() int {
  return file.Number
}

func (torrent TorrentListItem) Id() int {
  return torrent.TorrentId
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
    map[string]interface{} {
      "fields": []string{
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

func DeleteRequest(conn Connection, token string, ids []int, withData bool) TRequest {
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

func DetailsRequest(conn Connection, token string, id int, fields []string) TRequest {
  return TRequest{
    conn,
    "torrent-get",
    token,
    map[string]interface{} { "ids": []int{ id }, "fields": fields }}
}

func SetPriorityRequest(conn Connection, token string, id int, files []int, priority int) TRequest {
  var priorityValue string
  switch priority {
  case TR_PRIORITY_NORMAL:
    priorityValue = "priority-normal"
  case TR_PRIORITY_HIGH:
    priorityValue = "priority-high"
  case TR_PRIORITY_LOW:
    priorityValue = "priority-low"
  }

  return TRequest{
    conn,
    "torrent-set",
    token,
    map[string]interface{}{
      "ids": []int{ id },
      priorityValue: files}}
}

func SetWantedRequest(conn Connection, token string, id int, files []int, wanted bool) TRequest {
  var wantedValue string
  switch wanted {
  case true:
    wantedValue = "files-wanted"
  case false:
    wantedValue = "files-unwanted"
  }

  return TRequest{
    conn,
    "torrent-set",
    token,
    map[string]interface{}{
      "ids": []int{ id },
      wantedValue: files}}
}

func SetDownloadLimitRequest(conn Connection, token string, id int, value int) TRequest {
  var limited bool
  if value > 0 {
    limited = true
  }

  return TRequest{
    conn,
    "torrent-set",
    token,
    map[string]interface{}{
      "ids": []int{ id },
      "downloadLimit": value,
      "downloadLimited": limited}}
}

func SetUploadLimitRequest(conn Connection, token string, id int, value int) TRequest {
  var limited bool
  if value > 0 {
    limited = true
  }

  return TRequest{
    conn,
    "torrent-set",
    token,
    map[string]interface{}{
      "ids": []int{ id },
      "uploadLimit": value,
      "uploadLimited": limited}}
}

func StartTorrentRequest(conn Connection, token string, ids []int) TRequest {
  return TRequest{
    conn,
    "torrent-start",
    token,
    map[string]interface{}{
      "ids": ids}}
}

func StopTorrentRequest(conn Connection, token string, ids []int) TRequest {
  return TRequest{
    conn,
    "torrent-stop",
    token,
    map[string]interface{}{
      "ids": ids}}
}

func GetSessionSettingsRequest(conn Connection, token string) TRequest {
  return TRequest{
    conn,
    "session-get",
    token,
    map[string]interface{}{
      "fields": []string{
        "speed-limit-up",
        "speed-limit-up-enabled",
        "speed-limit-down",
        "speed-limit-down-enabled"}}}
}

func SetGlobalUploadLimitRequest(conn Connection, token string, value int) TRequest {
  var limited bool
  if value > 0 {
    limited = true
  }

  return TRequest{
    conn,
    "session-set",
    token,
    map[string]interface{}{
      "speed-limit-up": value,
      "speed-limit-up-enabled": limited}}
}

func SetGlobalDownloadLimitRequest(conn Connection, token string, value int) TRequest {
  var limited bool
  if value > 0 {
    limited = true
  }

  return TRequest{
    conn,
    "session-set",
    token,
    map[string]interface{}{
      "speed-limit-down": value,
      "speed-limit-down-enabled": limited}}
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

func (client *Client) List() (*[]TorrentListItem, error) {
  body, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return ListRequest(conn, token).ToRequest()
  })

  if err != nil {
    return nil, err
  }

  var listResponse TorrentListResponse
  json.Unmarshal(body, &listResponse)

  return &listResponse.Arguments.Torrents, err
}

func (client *Client) Delete(ids []int, withData bool) error {
  _, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return DeleteRequest(conn, token, ids, withData).ToRequest()
  })

  return err
}

func (client *Client) AddTorrent(url string, path string) (error) {
  body, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return AddRequest(conn, token, url, path, false).ToRequest()
  })

  if err != nil {
    return err
  }

  var response TorrentAddResponse
  jsonErr := json.Unmarshal(body, &response)
  if jsonErr != nil {
    return fmt.Errorf("Error: %s", jsonErr)
  }

  if response.Arguments.Torrent == nil {
    return fmt.Errorf("%s", response.Result)
  }

  return nil
}

func (client *Client) TorrentDetails(id int) (*TorrentDetails, error) {
  fields := []string{
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
    "uploadRatio",
    "downloadLimit",
    "downloadLimited",
    "uploadLimit",
    "uploadLimited",
    "files",
    "fileStats"}

  body, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return DetailsRequest(conn, token, id, fields).ToRequest()
  })

  if err != nil {
    return nil, err
  }

  var response TorrentDetailsResponse
  jsonErr := json.Unmarshal(body, &response)
  if jsonErr != nil {
    return nil, fmt.Errorf("Error: %s", jsonErr)
  }

  if response.Arguments.Torrents == nil {
    return nil, fmt.Errorf("%s", response.Result)
  }

  if len(*response.Arguments.Torrents) == 0 {
    return nil, nil
  }

  internalTorrent := (*response.Arguments.Torrents)[0]
  files := make([]TorrentFile, len(*internalTorrent.Files))
  for index, file := range (*internalTorrent.Files) {
    files[index] = TorrentFile{
      index,
      file.BytesCompleted,
      file.Length,
      file.Name,
      (*internalTorrent.FileStats)[index].Wanted,
      (*internalTorrent.FileStats)[index].Priority}
  }

  torrent := TorrentDetails{
    internalTorrent.Id,
    internalTorrent.Name,
    internalTorrent.UploadSpeed,
    internalTorrent.DownloadSpeed,
    internalTorrent.Ratio,
    internalTorrent.Eta,
    internalTorrent.SizeWhenDone,
    internalTorrent.LeftUntilDone,
    internalTorrent.Status,
    internalTorrent.DownloadLimit,
    internalTorrent.DownloadLimited,
    internalTorrent.UploadLimit,
    internalTorrent.UploadLimited,
    files}

  return &torrent, nil
}

func (client *Client) SetPriority(id int, files []int, priority int) error {
  body, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return SetPriorityRequest(conn, token, id, files, priority).ToRequest()
  })

  if err != nil {
    return err
  }

  var response TorrentListResponse
  jsonErr := json.Unmarshal(body, &response)
  if jsonErr != nil {
    return fmt.Errorf("Error: %s", jsonErr)
  }

  if (response.Result != "success") {
    return fmt.Errorf("Error: %s", response.Result)
  }

  return nil
}

func (client *Client) SetWanted(id int, files []int, wanted bool) error {
  body, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return SetWantedRequest(conn, token, id, files, wanted).ToRequest()
  })

  if err != nil {
    return err
  }

  var response TorrentListResponse
  jsonErr := json.Unmarshal(body, &response)
  if jsonErr != nil {
    return fmt.Errorf("Error: %s", jsonErr)
  }

  if (response.Result != "success") {
    return fmt.Errorf("Error: %s", response.Result)
  }

  return nil
}

func (client *Client) SetDownloadLimit(id int, limit int) error {
  body, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return SetDownloadLimitRequest(conn, token, id, limit).ToRequest()
  })

  if err != nil {
    return err
  }

  var response TorrentListResponse
  jsonErr := json.Unmarshal(body, &response)
  if jsonErr != nil {
    return fmt.Errorf("Error: %s", jsonErr)
  }

  if (response.Result != "success") {
    return fmt.Errorf("Error: %s", response.Result)
  }

  return nil
}

func (client *Client) SetUploadLimit(id int, limit int) error {
  body, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return SetUploadLimitRequest(conn, token, id, limit).ToRequest()
  })

  if err != nil {
    return err
  }

  var response TorrentListResponse
  jsonErr := json.Unmarshal(body, &response)
  if jsonErr != nil {
    return fmt.Errorf("Error: %s", jsonErr)
  }

  if (response.Result != "success") {
    return fmt.Errorf("Error: %s", response.Result)
  }

  return nil
}

func (client *Client) UpdateActive(ids []int, active bool) error {
  body, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    if active {
      return StartTorrentRequest(conn, token, ids).ToRequest()
    } else {
      return StopTorrentRequest(conn, token, ids).ToRequest()
    }
  })

  if err != nil {
    return err
  }

  var response TorrentListResponse
  jsonErr := json.Unmarshal(body, &response)
  if jsonErr != nil {
    return fmt.Errorf("Error: %s", jsonErr)
  }

  if (response.Result != "success") {
    return fmt.Errorf("Error: %s", response.Result)
  }

  return nil
}

func (client *Client) SetGlobalUploadLimit(limit int) error {
  body, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return SetGlobalUploadLimitRequest(conn, token, limit).ToRequest()
  })

  if err != nil {
    return err
  }

  var response SetSessionSettingsResponse
  jsonErr := json.Unmarshal(body, &response)
  if jsonErr != nil {
    return fmt.Errorf("Error: %s", jsonErr)
  }

  if (response.Result != "success") {
    return fmt.Errorf("Error: %s", response.Result)
  }

  return nil
}

func (client *Client) SetGlobalDownloadLimit(limit int) error {
  body, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return SetGlobalDownloadLimitRequest(conn, token, limit).ToRequest()
  })

  if err != nil {
    return err
  }

  var response SetSessionSettingsResponse
  jsonErr := json.Unmarshal(body, &response)
  if jsonErr != nil {
    return fmt.Errorf("Error: %s", jsonErr)
  }

  if (response.Result != "success") {
    return fmt.Errorf("Error: %s", response.Result)
  }

  return nil
}

func (client *Client) GetSessionSettings() (*SessionSettings, error) {
  body, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return GetSessionSettingsRequest(conn, token).ToRequest()
  })

  if err != nil {
    return nil, err
  }

  var response SessionSettingsResponse
  jsonErr := json.Unmarshal(body, &response)
  if jsonErr != nil {
    return nil, fmt.Errorf("Error: %s", jsonErr)
  }

  if (response.Result != "success") {
    return nil, fmt.Errorf("Error: %s", response.Result)
  }

  return response.Arguments, nil
}
