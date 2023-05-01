package transmission

import "net/http"

func AddRequest(filename string, downloadDir string, paused bool) RequestBuilder {
	return func(conn Connection, token string)(*http.Request, error) {
		return TRequest{
			conn,
			"torrent-add",
			token,
			map[string]interface{} { "filename": filename, "download-dir": downloadDir, "paused": paused }}.ToRequest()
	}
}

type TorrentAddedInfo struct {
	HashString string `json:"hashString"`
	Id int						`json:"id"`
	Name string				`json:"name"`
}

type TorrentAddResponseArguments struct {
	Torrent *TorrentAddedInfo `json:"torrent-added"`
}

type TorrentAddResponse struct {
	ResultValue string												 `json:"result"`
	TagValue string														 `json:"tag"`
	ArgumentsValue TorrentAddResponseArguments `json:"arguments"`
}

func (response TorrentAddResponse) Result() string {
	return response.ResultValue
}

func (response TorrentAddResponse) Tag() string {
	return response.TagValue
}

func (response TorrentAddResponse) Arguments() interface{} {
	return response.ArgumentsValue
}
