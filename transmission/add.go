package transmission

type TorrentAddedInfo struct {
  HashString string `json:"hashString"`
  Id int            `json:"id"`
  Name string       `json:"name"`
}

type TorrentAddResponseArguments struct {
  Torrent *TorrentAddedInfo `json:"torrent-added"`
}

type TorrentAddResponse struct {
  ResultValue string                         `json:"result"`
  TagValue string                            `json:"tag"`
  ArgumentsValue TorrentAddResponseArguments `json:"arguments"`
}

func AddRequest(conn Connection, token string, filename string, downloadDir string, paused bool) TRequest {
  return TRequest{
    conn,
    "torrent-add",
    token,
    map[string]interface{} { "filename": filename, "download-dir": downloadDir, "paused": paused }}
}

func (response TorrentAddResponse) Result() string {
  return response.ResultValue
}

func (response TorrentAddResponse) Tag() string {
  return response.TagValue
}

func (response TorrentAddResponse) Arguments() interface{} {
  return response.ArgumentsValue
}
