package transmission

import "net/http"

func SetDownloadLimitRequest(id int, value int) RequestBuilder {
  var limited bool
  if value > 0 {
    limited = true
  }

  return func(conn Connection, token string) (*http.Request, error) {
    return TRequest{
      conn,
      "torrent-set",
      token,
      map[string]interface{}{
        "ids": []int{ id },
        "downloadLimit": value,
        "downloadLimited": limited}}.ToRequest()
  }
}

func SetUploadLimitRequest(id int, value int) RequestBuilder {
  var limited bool
  if value > 0 {
    limited = true
  }

  return func(conn Connection, token string) (*http.Request, error) {
    return TRequest{
      conn,
      "torrent-set",
      token,
      map[string]interface{}{
        "ids": []int{ id },
        "uploadLimit": value,
        "uploadLimited": limited}}.ToRequest()
  }
}
