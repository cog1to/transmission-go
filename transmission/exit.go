package transmission

import "net/http"

func ExitRequest() RequestBuilder {
	return func(conn Connection, token string)(*http.Request, error) {
		return TRequest{
			conn,
			"session-close",
			token,
			nil}.ToRequest()
	}
}

