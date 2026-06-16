package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator membungkus *validator.Validate dari go-playground/validator.
type Validator struct {
	validate *validator.Validate
}

// New membuat instance Validator baru.
func New() *Validator {
    v := validator.New()

    // Daftarkan fungsi untuk ambil nama field dari JSON tag
    // supaya error key pakai "product_name" bukan "ProductName"
    v.RegisterTagNameFunc(func(fld reflect.StructField) string {
        name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
        if name == "-" {
            return ""
        }
        return name
    })

    return &Validator{validate: v}
}

// Validate menjalankan validasi terhadap struct yang diberikan.
//
// Mengembalikan nil jika tidak ada error. Jika ada error, mengembalikan
// map[string]string dengan key berupa nama field dalam format JSON tag
// (snake_case) dan value berupa pesan error berbahasa Indonesia.
func (v *Validator) Validate(i interface{}) map[string]string {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	// Jika error bukan ValidationErrors (mis. input bukan struct),
	// kembalikan error umum.
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string]string{"error": "tidak valid"}
	}

	errs := make(map[string]string)
	typ := reflectStructType(i)

	for _, fe := range validationErrs {
		field := jsonFieldName(typ, fe.StructField())
		errs[field] = messageFor(fe)
	}

	return errs
}

// reflectStructType mengembalikan reflect.Type dari struct, dereference
// pointer bila perlu.
func reflectStructType(i interface{}) reflect.Type {
	t := reflect.TypeOf(i)
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// jsonFieldName mencari nama field dari JSON tag berdasarkan nama field Go.
// Jika tag json tidak ada, fallback ke nama field Go apa adanya.
func jsonFieldName(t reflect.Type, structField string) string {
	if t == nil || t.Kind() != reflect.Struct {
		return structField
	}

	sf, ok := t.FieldByName(structField)
	if !ok {
		return structField
	}

	tag := sf.Tag.Get("json")
	if tag == "" || tag == "-" {
		return structField
	}

	// Buang opsi tambahan seperti ",omitempty".
	name := strings.Split(tag, ",")[0]
	if name == "" {
		return structField
	}
	return name
}

// messageFor mengembalikan pesan error berbahasa Indonesia sesuai rule
// validasi yang gagal.
func messageFor(fe validator.FieldError) string {
	param := fe.Param()

	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "min":
		if isStringKind(fe.Kind()) {
			return fmt.Sprintf("minimal %s karakter", param)
		}
		return fmt.Sprintf("minimal %s", param)
	case "max":
		if isStringKind(fe.Kind()) {
			return fmt.Sprintf("maksimal %s karakter", param)
		}
		return fmt.Sprintf("maksimal %s", param)
	case "email":
		return "format email tidak valid"
	case "oneof":
		return fmt.Sprintf("harus salah satu dari: %s", param)
	case "uuid4":
		return "format UUID tidak valid"
	case "len":
		return fmt.Sprintf("harus tepat %s karakter", param)
	case "numeric":
		return "harus berupa angka"
	default:
		return "tidak valid"
	}
}

// isStringKind memeriksa apakah kind dari field adalah string.
func isStringKind(k reflect.Kind) bool {
	return k == reflect.String
}
