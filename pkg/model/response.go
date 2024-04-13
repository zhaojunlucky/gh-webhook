package model

type ListResponse[T any] struct {
	EntryCount int `json:"entryCount"`
	Entries    []T `json:"entries"`
}

func NewListResponse[T any](entries []T) ListResponse[T] {
	return ListResponse[T]{
		EntryCount: len(entries),
		Entries:    entries,
	}
}

type ErrorMessage struct {
	Message string `json:"message"`
}

type ErrorMessageDTO struct {
	ErrorMessages []ErrorMessage `json:"errorMessages"`
}

func NewErrMsgs(errs ...string) []ErrorMessage {
	var errMsgs []ErrorMessage
	for _, err := range errs {
		errMsgs = append(errMsgs, ErrorMessage{
			Message: err,
		})
	}
	return errMsgs
}

func NewErrMsgsFromErr(errs ...error) []ErrorMessage {
	var errMsgs []ErrorMessage
	for _, err := range errs {
		errMsgs = append(errMsgs, ErrorMessage{
			Message: err.Error(),
		})
	}
	return errMsgs
}

func NewErrorMsgDTO(message ...string) ErrorMessageDTO {
	return ErrorMessageDTO{ErrorMessages: NewErrMsgs(message...)}
}

func NewErrorMsgDTOFromErr(err ...error) ErrorMessageDTO {
	return ErrorMessageDTO{ErrorMessages: NewErrMsgs()}
}

func NewErrorMsgDTOFromError(err ...error) ErrorMessageDTO {
	return ErrorMessageDTO{
		ErrorMessages: NewErrMsgsFromErr(err...),
	}
}

type IDResponse struct {
	ID uint `json:"id"`
}

func NewIDResponse(id uint) IDResponse {
	return IDResponse{ID: id}
}
