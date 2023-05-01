package suggestions

import (
	"os"
	"strings"
	"sort"
	"utils"
)

func GetSuggestedDirs(input string) []string {
	return getSuggestedPaths(input, false)
}

func GetSuggestedFiles(input string) []string {
	return getSuggestedPaths(input, true)
}

/** Private methods **/

func getSuggestedPaths(input string, includeFiles bool) []string {
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

	var expandedBase string
	if strings.HasPrefix(base, "~/") {
		expandedBase = utils.ExpandHome(base)
	} else {
		expandedBase = base
	}

	file, err := os.Open(expandedBase)
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
		matching = filter(matching, func(file os.FileInfo)(bool) { return strings.HasPrefix(file.Name(), name) })
	}

	if len(name) == 0 || !strings.HasPrefix(name, ".") {
		matching = filter(
			matching,
			func(file os.FileInfo)(bool) {
				return !strings.HasPrefix(file.Name(), ".") && (includeFiles || file.IsDir())
			})
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
