package transmission

import (
  "net/http"
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
  err := client.performJson(
    func(conn Connection, token string)(*http.Request, error) {
      return ListRequest(conn, token).ToRequest()
    },
    &response)

  if err != nil {
    return nil, err
  }

  args := response.Arguments().(TorrentListResponseArguments)
  return &args.Torrents, err
}

func (client *Client) Delete(ids []int, withData bool) error {
  _, err := client.perform(func(conn Connection, token string)(*http.Request, error) {
    return DeleteRequest(conn, token, ids, withData).ToRequest()
  })

  return err
}

func (client *Client) AddTorrent(url string, path string) (error) {
  var response TorrentAddResponse
  err := client.performJson(
    func(conn Connection, token string)(*http.Request, error) {
      return AddRequest(conn, token, url, path, false).ToRequest()
    },
    &response)

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
    "fileStats"}

  var response TorrentDetailsResponse
  err := client.performJson(
    func(conn Connection, token string)(*http.Request, error) {
      return DetailsRequest(conn, token, id, fields).ToRequest()
    },
    &response)

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
    files}

  return &torrent, nil
}

func (client *Client) SetPriority(id int, files []int, priority int) error {
  return client.performWithoutData(func(conn Connection, token string)(*http.Request, error) {
    return SetPriorityRequest(conn, token, id, files, priority).ToRequest()
  })
}

func (client *Client) SetWanted(id int, files []int, wanted bool) error {
  return client.performWithoutData(func(conn Connection, token string)(*http.Request, error) {
    return SetWantedRequest(conn, token, id, files, wanted).ToRequest()
  })
}

func (client *Client) SetDownloadLimit(id int, limit int) error {
  return client.performWithoutData(func(conn Connection, token string)(*http.Request, error) {
    return SetDownloadLimitRequest(conn, token, id, limit).ToRequest()
  })
}

func (client *Client) SetUploadLimit(id int, limit int) error {
  return client.performWithoutData(func(conn Connection, token string)(*http.Request, error) {
    return SetUploadLimitRequest(conn, token, id, limit).ToRequest()
  })
}

func (client *Client) UpdateActive(ids []int, active bool) error {
  return client.performWithoutData(func(conn Connection, token string)(*http.Request, error) {
    if active {
      return StartTorrentRequest(conn, token, ids).ToRequest()
    } else {
      return StopTorrentRequest(conn, token, ids).ToRequest()
    }
  })
}

func (client *Client) SetGlobalUploadLimit(limit int) error {
  return client.performWithoutData(func(conn Connection, token string)(*http.Request, error) {
    return SetGlobalUploadLimitRequest(conn, token, limit).ToRequest()
  })
}

func (client *Client) SetGlobalDownloadLimit(limit int) error {
  return client.performWithoutData(func(conn Connection, token string)(*http.Request, error) {
    return SetGlobalDownloadLimitRequest(conn, token, limit).ToRequest()
  })
}

func (client *Client) GetSessionSettings() (*SessionSettings, error) {
  var response SessionSettingsResponse
  err := client.performJson(
    func(conn Connection, token string)(*http.Request, error) {
      return GetSessionSettingsRequest(conn, token).ToRequest()
    },
    &response)

  if err != nil {
    return nil, err
  }

  args := response.Arguments().(*SessionSettings)
  return args, nil
}

