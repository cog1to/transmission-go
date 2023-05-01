package transmission

import "net/http"

func SetGlobalUploadLimitRequest(value int) RequestBuilder {
	var limited bool
	if value > 0 {
		limited = true
	}

	return func(conn Connection, token string) (*http.Request, error) {
		return TRequest{
			conn,
			"session-set",
			token,
			map[string]interface{}{
				"speed-limit-up": value,
				"speed-limit-up-enabled": limited}}.ToRequest()
	}
}

func SetGlobalDownloadLimitRequest(value int) RequestBuilder {
	var limited bool
	if value > 0 {
		limited = true
	}

	return func(conn Connection, token string) (*http.Request, error) {
		return TRequest{
			conn,
			"session-set",
			token,
			map[string]interface{}{
				"speed-limit-down": value,
				"speed-limit-down-enabled": limited}}.ToRequest()
	}
}

func UpdateActiveRequest(active bool, ids []int) RequestBuilder {
	var command string
	if active {
		command = "torrent-start"
	} else {
		command = "torrent-stop"
	}

	return func(conn Connection, token string) (*http.Request, error) {
		return TRequest{
			conn,
			command,
			token,
			map[string]interface{}{
				"ids": ids}}.ToRequest()
	}
}
