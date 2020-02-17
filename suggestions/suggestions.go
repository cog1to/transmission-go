package suggestions

import (
  "os"
  "strings"
  "sort"
)

func GetSuggestedPaths(input string) []string {
  if len(input) == 0 {
    return []string{}
  }

  var base, name string = input, ""
  if delimiterIndex := strings.LastIndex(input, "/"); delimiterIndex >= 0 {
    base = input[:delimiterIndex]
    if len(input) > (delimiterIndex+1) {
      name = input[delimiterIndex+1:]
    }
  }

  if !strings.HasSuffix(base, "/") {
    base = base + "/"
  }

  file, err := os.Open(base)
  if err != nil || file == nil {
    return []string{}
  }

  stats, err := file.Stat()
  if err != nil || !stats.IsDir() {
    return []string{}
  }

  files, err := file.Readdir(0)
  if err != nil || len(files) == 0 {
    return []string{}
  }

  matching := files
  if len(name) > 0 {
    matching = filter(files, func(file os.FileInfo)(bool) { return strings.HasPrefix(file.Name(), name) })
  }

  result := apply(matching, func(file os.FileInfo)(string) {
    suggestion := base + file.Name()
    if (file.IsDir()) {
      suggestion += "/"
    }
    return suggestion
  })
  sort.Strings(result)
  return result
}

func filter(input []os.FileInfo, match func(os.FileInfo)(bool)) []os.FileInfo {
  output := make([]os.FileInfo, 0, len(input))
  for _, item := range input {
    if match(item) {
      output = append(output, item)
    }
  }
  return output
}

func apply(input []os.FileInfo, operator func(os.FileInfo)(string)) []string {
  output := make([]string, len(input))
  for index, item := range input {
    output[index] = operator(item)
  }
  return output
}
