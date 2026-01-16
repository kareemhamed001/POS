package validation

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

func EgyptianPhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	re := regexp.MustCompile(`^(?:\+20|0020)?1[0125][0-9]{8}$`)
	return re.MatchString(phone)
}

func NewValidator() *validator.Validate {
	v := validator.New()
	_ = v.RegisterValidation("egyptianphone", EgyptianPhone)
	return v
}

func FormatValidationError(err error) string {
	if verrs, ok := err.(validator.ValidationErrors); ok {
		msgs := make([]string, 0, len(verrs))
		for _, fe := range verrs {
			msgs = append(msgs, fe.Field()+" "+fe.Tag())
		}
		return strings.Join(msgs, ", ")
	}

	return "invalid request payload"
}
