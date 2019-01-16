package structure

type Identity32 struct {
	Id int32 `json:"id" valid:"required~Required"`
}

type Identity64 struct {
	Id int64 `json:"id" valid:"required~Required"`
}
