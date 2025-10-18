package transmission

import "net/http"

func VerifyRequest(ids []int) RequestBuilder {
	return func(conn Connection, token string)(*http.Request, error) {
		return TRequest{
			conn,
			"torrent-verify",
			token,
			map[string]interface{} { "ids": ids }}.ToRequest()
	}
}

