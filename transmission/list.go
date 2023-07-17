package transmission

import "net/http"

func ListRequest(conn Connection, token string) (*http.Request, error) {
	return TRequest{
		conn,
		"torrent-get",
		token,
		map[string]interface{} {
			"fields": []string{
				"error",
				"errorString",
				"eta",
				"id",
				"leftUntilDone",
				"name",
				"rateDownload",
				"rateUpload",
				"sizeWhenDone",
				"status",
				"downloadDir",
				"uploadRatio",
				"addedDate"}}}.ToRequest()
}

type TorrentListItem struct {
	TorrentId int					`json:"id"`
	Name string						`json:"name"`
	UploadSpeed float32		`json:"rateUpload"`
	DownloadSpeed float32 `json:"rateDownload"`
	Ratio float32					`json:"uploadRatio"`
	Eta int32							`json:"eta"`
	SizeWhenDone int64		`json:"sizeWhenDone"`
	LeftUntilDone int64		`json:"leftUntilDone"`
	Status int8						`json:"status"`
	DownloadDir string		`json:"downloadDir"`
	AddedDate int					`json:"addedDate"`
}

type TorrentListResponseArguments struct {
	Torrents []TorrentListItem `json:"torrents"`
}

type TorrentListResponse struct {
	ResultValue string													`json:"result"`
	TagValue string															`json:"tag"`
	ArgumentsValue TorrentListResponseArguments `json:"arguments"`
}

func (response TorrentListResponse) Result() string {
	return response.ResultValue
}

func (response TorrentListResponse) Tag() string {
	return response.TagValue
}

func (response TorrentListResponse) Arguments() interface{} {
	return response.ArgumentsValue
}
