package glerror

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Error struct {
	err     error
	code    int
	message string
}

func NewError(err error, code int, message string) *Error {
	return &Error{
		err:     err,
		code:    code,
		message: message,
	}
}

func (se Error) Error() error {
	return se.err
}

func (se Error) Code() int {
	return se.code
}

func (se Error) Message() string {
	return se.message
}

func (se Error) StructedError() gin.H {
	return gin.H{
		"error":   se.err.Error(),
		"code":    se.code,
		"message": se.message,
	}
}

func InternalServerError() (int, gin.H) {
	return http.StatusInternalServerError, NewError(errors.New("internal server error"), http.StatusInternalServerError, "Internal server error.").StructedError()
}

func InvalidIdFormatError() (int, gin.H) {
	return http.StatusBadRequest, NewError(errors.New("invalid id format"), http.StatusBadRequest, "Invalid ID format.").StructedError()
}

func ResourceNotFoundError() (int, gin.H) {
	return http.StatusNotFound, NewError(errors.New("resource not found"), http.StatusNotFound, "Resource not found.").StructedError()
}

func UnauthorizedError() (int, gin.H) {
	return http.StatusUnauthorized, NewError(errors.New("not authorized"), http.StatusUnauthorized, "Not authorized.").StructedError()
}

func ResourceAlreadyExistError() (int, gin.H) {
	return http.StatusBadRequest, NewError(errors.New("resource already exist"), http.StatusBadRequest, "Resource already exist.").StructedError()
}

func BadRequestError() (int, gin.H) {
	return http.StatusBadRequest, NewError(errors.New("bad request"), http.StatusBadRequest, "Bad request.").StructedError()
}

func ConflictError() (int, gin.H) {
	return http.StatusConflict, NewError(errors.New("resource already exist"), http.StatusConflict, "resource already exist.").StructedError()
}

func InvalidBodyError(err error) (int, gin.H) {
	return http.StatusBadRequest, NewError(err, http.StatusBadRequest, "Bind error for request. See server logs").StructedError()
}

func InvalidBodyData(err error) (int, gin.H) {
	return http.StatusBadRequest, NewError(err, http.StatusBadRequest, "data struct validation error for request. See server logs").StructedError()
}

var (
	ErrUnauthorizedDataAccess = errors.New("unauthorized data access error")
)
