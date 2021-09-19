package transmission

import (
  "fmt"
)

/* Data */

const (
  TR_ETA_NOT_AVAIL = -1
  TR_ETA_UNKNOWN = -2
)

const (
  TR_STATUS_STOPPED = 0       /* Torrent is stopped */
  TR_STATUS_CHECK_WAIT = 1    /* Queued to check files */
  TR_STATUS_CHECK = 2         /* Checking files */
  TR_STATUS_DOWNLOAD_WAIT = 3 /* Queued to download */
  TR_STATUS_DOWNLOAD = 4      /* Downloading */
  TR_STATUS_SEED_WAIT = 5     /* Queued to seed */
  TR_STATUS_SEED = 6          /* Seeding */
)

const (
  TR_PRIORITY_NORMAL = 0
  TR_PRIORITY_HIGH = 1
  TR_PRIORITY_LOW = -1
)

/* Helpers */

func (file TorrentFile) Id() int {
  return file.Number
}

func (torrent TorrentListItem) Id() int {
  return torrent.TorrentId
}

/* Requests */

func (client *Client) List() (*[]TorrentListItem, error) {
  var response TorrentListResponse
  err := client.performJson(ListRequest, &response)

  if err != nil {
    return nil, err
  }

  args := response.Arguments().(TorrentListResponseArguments)
  return &args.Torrents, err
}

func (client *Client) Delete(ids []int, withData bool) error {
  _, err := client.perform(DeleteRequest(ids, withData))

  return err
}

func (client *Client) AddTorrent(url string, path string) (error) {
  var response TorrentAddResponse
  err := client.performJson(AddRequest(url, path, false), &response)

  if err != nil {
    return err
  }

  args := response.Arguments().(TorrentAddResponseArguments)
  if (args.Torrent == nil) {
    return fmt.Errorf("Error: %s", response.Result())
  } else {
    return nil
  }
}

func (client *Client) TorrentDetails(id int) (*TorrentDetails, error) {
  fields := []string{
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
    "uploadRatio",
    "downloadLimit",
    "downloadLimited",
    "uploadLimit",
    "uploadLimited",
    "files",
    "downloadDir",
    "fileStats"}

  var response TorrentDetailsResponse
  err := client.performJson(DetailsRequest(id, fields), &response)

  if err != nil {
    return nil, err
  }

  args := response.Arguments().(TorrentDetailsResponseArguments)
  if args.Torrents == nil {
    return nil, fmt.Errorf("%s", response.Result())
  }

  if len(*args.Torrents) == 0 {
    return nil, nil
  }

  internalTorrent := (*args.Torrents)[0]
  files := make([]TorrentFile, len(*internalTorrent.Files))
  for index, file := range (*internalTorrent.Files) {
    files[index] = TorrentFile{
      index,
      file.BytesCompleted,
      file.Length,
      file.Name,
      (*internalTorrent.FileStats)[index].Wanted,
      (*internalTorrent.FileStats)[index].Priority}
  }

  torrent := TorrentDetails{
    internalTorrent.Id,
    internalTorrent.Name,
    internalTorrent.UploadSpeed,
    internalTorrent.DownloadSpeed,
    internalTorrent.Ratio,
    internalTorrent.Eta,
    internalTorrent.SizeWhenDone,
    internalTorrent.LeftUntilDone,
    internalTorrent.Status,
    internalTorrent.DownloadLimit,
    internalTorrent.DownloadLimited,
    internalTorrent.UploadLimit,
    internalTorrent.UploadLimited,
    internalTorrent.DownloadDir,
    files}

  return &torrent, nil
}

func (client *Client) SetPriority(id int, files []int, priority int) error {
  return client.performWithoutData(SetPriorityRequest(id, files, priority))
}

func (client *Client) SetWanted(id int, files []int, wanted bool) error {
  return client.performWithoutData(SetWantedRequest(id, files, wanted))
}

func (client *Client) SetDownloadLimit(id int, limit int) error {
  return client.performWithoutData(SetDownloadLimitRequest(id, limit))
}

func (client *Client) SetUploadLimit(id int, limit int) error {
  return client.performWithoutData(SetUploadLimitRequest(id, limit))
}

func (client *Client) SetLocation(ids []int, location string) error {
  return client.performWithoutData(SetLocationRequest(ids, location))
}

func (client *Client) UpdateActive(ids []int, active bool) error {
  return client.performWithoutData(UpdateActiveRequest(active, ids))
}

func (client *Client) SetGlobalUploadLimit(limit int) error {
  return client.performWithoutData(SetGlobalUploadLimitRequest(limit))
}

func (client *Client) SetGlobalDownloadLimit(limit int) error {
  return client.performWithoutData(SetGlobalDownloadLimitRequest(limit))
}

func (client *Client) GetSessionSettings() (*SessionSettings, error) {
  var response SessionSettingsResponse
  err := client.performJson(GetSessionSettingsRequest, &response)

  if err != nil {
    return nil, err
  }

  args := response.Arguments().(*SessionSettings)
  return args, nil
}

func (client *Client) Exit() error {
  return client.performWithoutData(ExitRequest())
}
