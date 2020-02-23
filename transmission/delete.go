package transmission

func DeleteRequest(conn Connection, token string, ids []int, withData bool) TRequest {
  return TRequest{
    conn,
    "torrent-remove",
    token,
    map[string]interface{} { "ids": ids, "delete-local-data": withData }}
}

