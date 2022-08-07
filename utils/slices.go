package utils

func Remove(slice []rune, s int) []rune {
  return append(slice[:s], slice[s+1:]...)
}

func RemoveInt(slice []int, el int) []int {
    for index, element := range slice {
    if element == el {
      return append(slice[:index], slice[index+1:]...)
    }
  }
  return slice
}

func Contains(slice []int, el int) bool {
  for _, element := range slice {
    if element == el {
      return true
    }
  }
  return false
}

func IndexOf(slice []string, el string) int {
  for ind, element := range slice {
    if element == el {
      return ind
    }
  }

  return -1
}
