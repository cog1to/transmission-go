package transmission

import "net/http"

func DetailsRequest(id int, fields []string) RequestBuilder {
  return func(conn Connection, token string) (*http.Request, error) {
    return TRequest{
      conn,
      "torrent-get",
      token,
      map[string]interface{} { "ids": []int{ id }, "fields": fields }}.ToRequest()
  }
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
  DownloadDir string                    `json:"downloadDir"`
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
  DownloadDir string
  Files []TorrentFile
}

type TorrentDetailsResponseArguments struct {
  Torrents *[]TorrentDetailsInternal `json:"torrents"`
}

type TorrentDetailsResponse struct {
  ResultValue string                             `json:"result"`
  TagValue string                                `json:"tag"`
  ArgumentsValue TorrentDetailsResponseArguments `json:"arguments"`
}

func (response TorrentDetailsResponse) Result() string {
  return response.ResultValue
}

func (response TorrentDetailsResponse) Tag() string {
  return response.TagValue
}

func (response TorrentDetailsResponse) Arguments() interface{} {
  return response.ArgumentsValue
}

