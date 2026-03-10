package response

import (
	"net/http"

	"learning-go/internal/application/i18n"

	"github.com/gin-gonic/gin"
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
		if len(err) == 1 {
			errPayload = err[0]
		} else {
			payload := make([]interface{}, 0, len(err))
			for _, e := range err {
				if v, ok := e.(error); ok {
					payload = append(payload, v.Error())
				} else {
					payload = append(payload, e)
				}
			}
			errPayload = payload
		}

		if v, ok := errPayload.(error); ok {
			errPayload = v.Error()
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

func InternalServerError(c *gin.Context, message string, err ...interface{}) {
	msg := message
	if msg == "" {
		msg = "common.internal_server_error"
	}
	Error(c, http.StatusInternalServerError, msg, err...)
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
