package transmission

type TResponse interface {
  Tag() string
  Result() string
  Arguments() interface{}
}
