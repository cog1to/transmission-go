package transmission

import "net/http"

func SetTrackerListRequest(id int, list []string) RequestBuilder {
	return func(conn Connection, token string) (*http.Request, error) {
		var listString string
		for _, item := range(list) {
			listString += item + "\n"
		}

		return TRequest{
			conn,
			"torrent-set",
			token,
			map[string]interface{}{
				"ids": []int{ id },
				"trackerList": listString,
			},
		}.ToRequest()
	}
}
