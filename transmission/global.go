package transmission

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

