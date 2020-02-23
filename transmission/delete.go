package transmission

import "net/http"

func DeleteRequest(ids []int, withData bool) RequestBuilder {
  return func(conn Connection, token string)(*http.Request, error) {
    return TRequest{
      conn,
      "torrent-remove",
      token,
      map[string]interface{} { "ids": ids, "delete-local-data": withData }}.ToRequest()
  }
}

