package tui

func Remove(attrs []Attribute, el Attribute) []Attribute {
    for index, element := range attrs {
    if element == el {
      return append(attrs[:index], attrs[index+1:]...)
    }
  }
  return attrs
}

func Contains(attrs []Attribute, el Attribute) bool {
  for _, element := range attrs {
    if element == el {
      return true
    }
  }
  return false
}

func Same(left []Attribute, right []Attribute) bool {
  if len(left) != len(right) {
    return false
  }

  for _, attr := range(left) {
    if !Contains(right, attr) {
      return false
    }
  }

  return true
}

func SameColor(left, right *colorPair) bool {
  if left == nil && right == nil {
    return true
  } else if left == nil || right == nil {
    return false
  } else {
    return left.front == right.front && left.back == right.back
  }
}

