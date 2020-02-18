package windows

import (
  "../transmission"
  "fmt"
)

func formatSize(size int64) string {
  switch true {
  case size >= (1024 * 1024 * 1024):
    return fmt.Sprintf("%.2fGB", float64(size)/float64(1024 * 1024 * 1024))
  case size >= (1024 * 1024):
    return fmt.Sprintf("%.2fMB", float64(size)/float64(1024 * 1024))
  case size >= 1024:
    return fmt.Sprintf("%.2fKB", float64(size)/float64(1024))
  default:
    return fmt.Sprintf("%dB", size)
  }
}

func formatSpeed(speed float32) string {
  switch true {
  case speed == 0:
    return "0"
  case speed >= (1024 * 1024 * 1024):
    return fmt.Sprintf("%.2fGB", speed/float32(1024 * 1024 * 1024))
  case speed >= (1024 * 1024):
    return fmt.Sprintf("%.2fMB", speed/float32(1024 * 1024))
  case speed >= 1024:
    return fmt.Sprintf("%.2fKB", speed/1024)
  default:
    return fmt.Sprintf("%.2fB", speed)
  }
}

func formatSpeedWithFlag(speed float32, honored bool) string {
  if honored {
    return formatSpeed(speed)
  } else {
    return "0"
  }
}

func formatStatus(status int8) string {
  switch status {
  case transmission.TR_STATUS_STOPPED:
    return "Stopped"
  case transmission.TR_STATUS_CHECK_WAIT:
    return "Check queue"
  case transmission.TR_STATUS_CHECK:
    return "Checking"
  case transmission.TR_STATUS_DOWNLOAD_WAIT:
    return "In Queue"
  case transmission.TR_STATUS_DOWNLOAD:
    return "Download"
  case transmission.TR_STATUS_SEED_WAIT:
    return "Seed queue"
  case transmission.TR_STATUS_SEED:
    return "Seeding"
  }
  return "Unknown"
}

func formatPriority(priority int) string {
  switch priority {
  case 0:
    return "Normal"
  case 1:
    return "High"
  case -1:
    return "Low"
  default:
    return "Unknown"
  }
}

func formatFlag(flag bool) string {
  if flag {
    return "Yes"
  } else {
    return "No"
  }
}

func formatTime(time int32, done bool) string {
  if time == -1 && done {
    return "Done"
  } else if time == -2 || (time == -1 && !done) {
    return "Unknown"
  } else {
    if time < 60 {
      return fmt.Sprintf("%ds", time)
    }
    if time < (60 * 60) {
      return fmt.Sprintf("%0.1fmin", float32(time) / 60.0)
    }
    if time < (60 * 60 * 24) {
      return fmt.Sprintf("%0.1fhr", float32(time) / 3600.0)
    } else {
      return fmt.Sprintf("%0.1fd", float32(time) / 86400.0)
    }
  }
}
