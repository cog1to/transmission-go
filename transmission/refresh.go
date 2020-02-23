package transmission

func RefreshRequest(conn Connection) TRequest {
  return TRequest{
    conn,
    "",
    "",
    nil}
}
