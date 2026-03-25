package response

import (
	"errors"
	"net/http"
	apperr "learning-go/internal/shared/error"
	"learning-go/internal/shared/i18n"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type APIResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
	Error    interface{} `json:"error,omitempty"`
}

func Success(c *gin.Context, status int, data interface{}, message ...string) {
	msg := "common.successfully"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	lang := getLang(c)
	msg = i18n.TranslateText(lang, msg)
	c.JSON(status, APIResponse{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

func SuccessWithMetadata(c *gin.Context, status int, data interface{}, metadata interface{}, message ...string) {
	msg := "common.successfully"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	lang := getLang(c)
	msg = i18n.TranslateText(lang, msg)
	c.JSON(status, APIResponse{
		Success:  true,
		Message:  msg,
		Data:     data,
		Metadata: metadata,
	})
}

func Error(c *gin.Context, status int, message string, err ...interface{}) {
	msg := message
	if msg == "" {
		msg = "common.error"
	}
	lang := getLang(c)
	msg = i18n.TranslateText(lang, msg)

	var errPayload interface{}
	if len(err) > 0 {
		converted := make([]interface{}, 0, len(err))
		for _, e := range err {
			if v, ok := e.(error); ok {
				converted = append(converted, v.Error())
			} else {
				converted = append(converted, e)
			}
		}
		if len(converted) == 1 {
			errPayload = converted[0]
		} else {
			errPayload = converted
		}
	}

	errPayload = translatePayload(lang, errPayload)

	c.JSON(status, APIResponse{
		Success: false,
		Message: msg,
		Error:   errPayload,
	})
}

func BadRequest(c *gin.Context, message string, err ...interface{}) {
	msg := message
	if msg == "" {
		msg = "common.bad_request"
	}
	Error(c, http.StatusBadRequest, msg, err...)
}

// ValidationError formats Gin binding/validation errors into field-level details (HTTP 422).
func ValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make(map[string]string, len(ve))
		for _, fe := range ve {
			fields[fe.Field()] = fe.Tag()
		}
		UnprocessableEntity(c, "common.validation_failed", fields)
		return
	}
	UnprocessableEntity(c, "")
}


func Unauthorized(c *gin.Context, message string, err ...interface{}) {
	msg := message
	if msg == "" {
		msg = "common.unauthorized"
	}
	Error(c, http.StatusUnauthorized, msg, err...)
}

func Forbidden(c *gin.Context, message string, err ...interface{}) {
	msg := message
	if msg == "" {
		msg = "common.forbidden"
	}
	Error(c, http.StatusForbidden, msg, err...)
}

func Conflict(c *gin.Context, message string, err ...interface{}) {
	msg := message
	if msg == "" {
		msg = "common.conflict"
	}
	Error(c, http.StatusConflict, msg, err...)
}

func NotFound(c *gin.Context, message string, err ...interface{}) {
	msg := message
	if msg == "" {
		msg = "common.not_found"
	}
	Error(c, http.StatusNotFound, msg, err...)
}

func ServiceUnavailable(c *gin.Context, message string, err ...interface{}) {
	msg := message
	if msg == "" {
		msg = "common.service_unavailable"
	}
	Error(c, http.StatusServiceUnavailable, msg, err...)
}

func UnprocessableEntity(c *gin.Context, message string, err ...interface{}) {
	msg := message
	if msg == "" {
		msg = "common.validation_failed"
	}
	Error(c, http.StatusUnprocessableEntity, msg, err...)
}

func InternalServerError(c *gin.Context, message string, err ...interface{}) {
	msg := message
	if msg == "" {
		msg = "common.internal_server_error"
	}
	Error(c, http.StatusInternalServerError, msg, err...)
}

var codeToStatus = map[apperr.Code]int{
	apperr.CodeBadRequest:          http.StatusBadRequest,
	apperr.CodeUnauthorized:        http.StatusUnauthorized,
	apperr.CodeForbidden:           http.StatusForbidden,
	apperr.CodeNotFound:            http.StatusNotFound,
	apperr.CodeConflict:            http.StatusConflict,
	apperr.CodeUnprocessableEntity: http.StatusUnprocessableEntity,
	apperr.CodeInternalServerError: http.StatusInternalServerError,
	apperr.CodeServiceUnavailable:  http.StatusServiceUnavailable,
}

// HandleError maps AppError to HTTP response.
func HandleError(c *gin.Context, err error) {
	var domErr *apperr.AppError
	if !errors.As(err, &domErr) {
		InternalServerError(c, "")
		return
	}

	status, ok := codeToStatus[domErr.Code()]
	if !ok {
		status = http.StatusInternalServerError
	}

	if data := domErr.Data(); data != nil {
		lang := getLang(c)
		msg := i18n.TranslateText(lang, domErr.Message())
		c.JSON(status, APIResponse{
			Success: false,
			Message: msg,
			Error:   data,
		})
		return
	}

	Error(c, status, domErr.Message())
}

func getLang(c *gin.Context) string {
	lang := c.GetString("lang")
	if lang != "" {
		return i18n.Normalize(lang)
	}
	return i18n.FromAcceptLanguage(c.GetHeader("Accept-Language"))
}

func translatePayload(lang string, payload interface{}) interface{} {
	switch v := payload.(type) {
	case nil:
		return nil
	case string:
		return i18n.TranslateText(lang, v)
	case []string:
		out := make([]string, 0, len(v))
		for _, s := range v {
			out = append(out, i18n.TranslateText(lang, s))
		}
		return out
	case []interface{}:
		out := make([]interface{}, 0, len(v))
		for _, it := range v {
			if s, ok := it.(string); ok {
				out = append(out, i18n.TranslateText(lang, s))
				continue
			}
			out = append(out, it)
		}
		return out
	default:
		return payload
	}
}
