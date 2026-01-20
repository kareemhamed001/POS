package validation

import (
	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	entity "github.com/kareemhamed001/POS/internal/core/domain"
)

func EgyptianPhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	re := regexp.MustCompile(`^(?:\+20|0020)?1[0125][0-9]{8}$`)
	return re.MatchString(phone)
}

func ValidDiscountType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // omitempty allows empty
	}
	return entity.DiscountType(value).IsValid()
}

func ValidDiscountTypePtr(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Check if field kind is pointer before calling IsNil
	if field.Kind() != reflect.Ptr {
		// If it's not a pointer, treat as regular string
		value := field.String()
		if value == "" {
			return true
		}
		return entity.DiscountType(value).IsValid()
	}

	// For pointer fields, check if nil is acceptable
	if field.IsNil() {
		return true
	}

	value := field.Elem().String()
	if value == "" {
		return true // omitempty allows empty
	}
	return entity.DiscountType(value).IsValid()
}

func NewValidator() *validator.Validate {
	v := validator.New()
	_ = v.RegisterValidation("egyptianphone", EgyptianPhone)
	_ = v.RegisterValidation("discount_type_valid", ValidDiscountType)
	_ = v.RegisterValidation("discount_type_valid_ptr", ValidDiscountTypePtr)
	_ = v.RegisterValidation("mimetype", MimeType)
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

// MimeType validates that a file's content type matches allowed values.
// Usage: `validate:"mimetype=image/jpeg,image/png"` or with wildcard `image/*`.
func MimeType(fl validator.FieldLevel) bool {
	// Parse allowed types from the tag parameter
	param := strings.TrimSpace(fl.Param())
	if param == "" {
		// No constraint provided; accept any type
		return true
	}
	// split by spaces and commas
	fields := strings.FieldsFunc(param, func(r rune) bool { return r == ' ' || r == ',' })
	allowed := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			allowed = append(allowed, f)
		}
	}

	// Handle nil pointer fields (allowed with omitempty)
	field := fl.Field()
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return true
	}

	// Helper to detect content type from a FileHeader by sniffing bytes
	detect := func(fh *multipart.FileHeader) (string, error) {
		// Try Content-Type header first
		if ct := fh.Header.Get("Content-Type"); ct != "" {
			return ct, nil
		}
		// Fallback to sniffing bytes
		f, err := fh.Open()
		if err != nil {
			// If we cannot open, return octet-stream instead of failing hard
			return "application/octet-stream", nil
		}
		defer f.Close()
		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		if n <= 0 {
			return "application/octet-stream", nil
		}
		return http.DetectContentType(buf[:n]), nil
	}

	// Helper to check if detected type is allowed (supports wildcard `type/*`)
	isAllowed := func(ct string) bool {
		for _, a := range allowed {
			if strings.HasSuffix(a, "/*") {
				prefix := strings.TrimSuffix(a, "/*")
				if strings.HasPrefix(ct, prefix+"/") {
					return true
				}
			} else if strings.EqualFold(a, ct) {
				return true
			}
		}
		return false
	}

	// Map common extensions to MIME types for final fallback
	extToMime := func(name string) string {
		lower := strings.ToLower(name)
		switch {
		case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
			return "image/jpeg"
		case strings.HasSuffix(lower, ".png"):
			return "image/png"
		case strings.HasSuffix(lower, ".gif"):
			return "image/gif"
		case strings.HasSuffix(lower, ".webp"):
			return "image/webp"
		default:
			return "application/octet-stream"
		}
	}

	// Support single and multiple files
	if fh, ok := field.Interface().(*multipart.FileHeader); ok {
		ct, _ := detect(fh)
		if isAllowed(ct) {
			return true
		}
		// Fallback: if wildcard image/* allowed and extension is image, accept
		if strings.HasPrefix(ct, "application/") {
			ct2 := extToMime(fh.Filename)
			return isAllowed(ct2)
		}
		return false
	}
	if fhs, ok := field.Interface().([]*multipart.FileHeader); ok {
		for _, fh := range fhs {
			if fh == nil {
				continue
			}
			ct, _ := detect(fh)
			if isAllowed(ct) {
				continue
			}
			if strings.HasPrefix(ct, "application/") {
				ct2 := extToMime(fh.Filename)
				if isAllowed(ct2) {
					continue
				}
			}
			return false
		}
		return true
	}

	// If not a FileHeader, treat as string content type for flexibility
	if s, ok := field.Interface().(string); ok {
		return isAllowed(s)
	}

	// Unsupported type
	return false
}
