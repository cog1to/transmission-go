package transmission

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

