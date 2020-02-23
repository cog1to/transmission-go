package transmission

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
