package transmission

import "net/http"

func GetSessionSettingsRequest(conn Connection, token string) (*http.Request, error) {
  return TRequest{
    conn,
    "session-get",
    token,
    map[string]interface{}{
      "fields": []string{
        "speed-limit-up",
        "speed-limit-up-enabled",
        "speed-limit-down",
        "speed-limit-down-enabled"}}}.ToRequest()
}

type SessionSettings struct {
  UploadSpeedLimit int           `json:"speed-limit-up"`
  UploadSpeedLimitEnabled bool   `json:"speed-limit-up-enabled"`
  DownloadSpeedLimit int         `json:"speed-limit-down"`
  DownloadSpeedLimitEnabled bool `json:"speed-limit-down-enabled"`
}

type SessionSettingsResponse struct {
  ResultValue string              `json:"result"`
  TagValue string                 `json:"tag"`
  ArgumentsValue *SessionSettings `json:"arguments"`
}

func (response SessionSettingsResponse) Result() string {
  return response.ResultValue
}

func (response SessionSettingsResponse) Tag() string {
  return response.TagValue
}

func (response SessionSettingsResponse) Arguments() interface{} {
  return response.ArgumentsValue
}
