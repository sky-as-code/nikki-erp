package validator

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/util"
	"go.bryk.io/pkg/errors"
)

func StartValidationFlow(startWith ...Validatable) *ValidationFlow {
	flow := ValidationFlow{}
	if len(startWith) > 0 {
		return flow.Start().Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = startWith[0].Validate()
			return nil
		})
	}
	return flow.Start()

}

type ValidationFlow struct {
	vErrs *ft.ValidationErrors
	err   error
	skip  bool
}

func (this *ValidationFlow) Start() *ValidationFlow {
	this.vErrs = util.ToPtr(ft.NewValidationErrors())
	return this
}

func (this *ValidationFlow) Step(fn func(vErrs *ft.ValidationErrors) error, ignoreValidationError ...bool) (out *ValidationFlow) {
	defer func() {
		out = this
		if e := recover(); e != nil {
			err, ok := e.(error)
			if ok {
				this.err = err
				return
			}
			this.err = errors.New(e)
		}
	}()

	if this.err != nil || this.skip {
		return
	}

	this.err = fn(this.vErrs)

	if this.vErrs.Count() > 0 && (len(ignoreValidationError) == 0 || !ignoreValidationError[0]) {
		this.skip = true
	}
	return
}

func (this *ValidationFlow) End() (ft.ValidationErrors, error) {
	return *this.vErrs, this.err
}

type Validatable interface {
	Validate() ft.ValidationErrors
}
