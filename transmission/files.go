package transmission

import "net/http"

func SetPriorityRequest(id int, files []int, priority int) RequestBuilder {
  var priorityValue string
  switch priority {
  case TR_PRIORITY_NORMAL:
    priorityValue = "priority-normal"
  case TR_PRIORITY_HIGH:
    priorityValue = "priority-high"
  case TR_PRIORITY_LOW:
    priorityValue = "priority-low"
  }

  return func(conn Connection, token string) (*http.Request, error) {
    return TRequest{
      conn,
      "torrent-set",
      token,
      map[string]interface{}{
        "ids": []int{ id },
        priorityValue: files}}.ToRequest()
  }
}

func SetWantedRequest(id int, files []int, wanted bool) RequestBuilder {
  var wantedValue string
  switch wanted {
  case true:
    wantedValue = "files-wanted"
  case false:
    wantedValue = "files-unwanted"
  }

  return func(conn Connection, token string) (*http.Request, error) {
    return TRequest{
      conn,
      "torrent-set",
      token,
      map[string]interface{}{
        "ids": []int{ id },
        wantedValue: files}}.ToRequest()
  }
}

