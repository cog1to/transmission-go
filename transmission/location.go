package transmission

import "net/http"

func SetLocationRequest(ids []int, value string) RequestBuilder {
	return func(conn Connection, token string) (*http.Request, error) {
		return TRequest{
			conn,
			"torrent-set-location",
			token,
			map[string]interface{}{
				"ids": ids,
				"location": value,
				"move": true}}.ToRequest()
	}
}
