package handlers

type SeaReponse[R any] struct {
	Error string `json:"error,omitempty"`
	Data  R      `json:"data,omitempty"`
}

func NewDataResponse[R any](data R) SeaReponse[R] {
	res := SeaReponse[R]{
		Data: data,
	}
	return res
}

func NewErrorResponse(err error) SeaReponse[any] {
	if err == nil {
		return SeaReponse[any]{
			Error: "Unknown error",
		}
	}
	res := SeaReponse[any]{
		Error: err.Error(),
	}
	return res
}
