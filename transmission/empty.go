package transmission

type GenericResponse struct {
  ResultValue string         `json:"result"`
  TagValue string            `json:"tag"`
  ArgumentsValue interface{} `json:"arguments"`
}

func (response GenericResponse) Result() string {
  return response.ResultValue
}

func (response GenericResponse) Tag() string {
  return response.TagValue
}

func (response GenericResponse) Arguments() interface{} {
  return response.ArgumentsValue
}
