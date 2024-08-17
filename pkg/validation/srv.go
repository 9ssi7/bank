package validation

import (
	"context"
	"errors"
	"strings"

	"github.com/9ssi7/bank/pkg/rescode"
	"github.com/9ssi7/bank/pkg/state"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/tr"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Srv struct {
	validator *validator.Validate
	uni       *ut.UniversalTranslator
}

func New() *Srv {
	v := validator.New()
	v.RegisterCustomTypeFunc(validateUUID, uuid.UUID{})
	_ = v.RegisterValidation("username", validateUserName)
	_ = v.RegisterValidation("password", validatePassword)
	_ = v.RegisterValidation("locale", validateLocale)
	_ = v.RegisterValidation("slug", validateSlug)
	_ = v.RegisterValidation("gender", validateGender)
	_ = v.RegisterValidation("phone", validatePhone)
	_ = v.RegisterValidation("currency", validateCurrency)
	_ = v.RegisterValidation("amount", validateAmount)
	return &Srv{validator: v, uni: ut.New(tr.New(), en.New())}
}

// ValidateStruct validates the given struct.
func (s *Srv) ValidateStruct(ctx context.Context, sc interface{}) error {
	var errs []*ErrorResponse
	translator := s.getTranslator(ctx)
	err := s.validator.StructCtx(ctx, sc)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			ns := s.mapStructNamespace(err.Namespace())
			if ns != "" {
				element.Namespace = ns
			}
			element.Field = err.Field()
			element.Value = err.Value()
			element.Message = err.Translate(translator)
			errs = append(errs, &element)
		}
	}
	if len(errs) > 0 {
		return rescode.ValidationFailed(errors.New("validation failed")).SetData(errs)
	}
	return nil
}

// ValidateMap validates the giveb struct.
func (s *Srv) ValidateMap(ctx context.Context, m map[string]interface{}, rules map[string]interface{}) error {
	var errs []*ErrorResponse
	errMap := s.validator.ValidateMapCtx(ctx, m, rules)
	translator := s.getTranslator(ctx)
	for key, err := range errMap {
		var element ErrorResponse
		if _err, ok := err.(validator.ValidationErrors); ok {
			for _, err := range _err {
				element.Namespace = err.Namespace()
				element.Field = err.Field()
				if element.Field == "" {
					element.Field = key
				}
				element.Value = err.Value()
				element.Message = err.Translate(translator)
				errs = append(errs, &element)
			}
			continue
		}
	}
	if len(errs) > 0 {
		return rescode.ValidationFailed(errors.New("validation failed")).SetData(errs)
	}
	return nil
}

func (s *Srv) getTranslator(ctx context.Context) ut.Translator {
	locale := state.GetLocale(ctx)
	translator, found := s.uni.GetTranslator(locale)
	if !found {
		translator = s.uni.GetFallback()
	}
	return translator
}

func (s *Srv) mapStructNamespace(ns string) string {
	str := strings.Split(ns, ".")
	return strings.Join(str[1:], ".")
}
