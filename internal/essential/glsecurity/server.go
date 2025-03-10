package glsecurity

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/guardlight/server/internal/essential/glerror"
	"go.uber.org/zap"
)

func ReuseBindAndValidate(c *gin.Context, obj any) error {
	if err := c.ShouldBindBodyWith(&obj, binding.JSON); err != nil {
		zap.S().Errorw("failed to bind to object to body", "error", err)
		c.JSON(glerror.InvalidBodyError(err))
		return err
	}

	return nil
}

func ValidateData(data interface{}) error {
	validatorInstance := validator.New(validator.WithRequiredStructEnabled())

	// Validate the configuration according to validation tags in the structs.
	if err := validatorInstance.Struct(data); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			zap.S().Errorw("request field error",
				"field", err.Field(),
				"value", err.Value(),
				"validation_type", err.Tag(),
				"field_type", err.Type(),
			)
		}
		zap.S().Debugw("request error", "error", err)
		return err
	}
	return nil
}
