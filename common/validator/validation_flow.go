package validator

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/util"
	"go.bryk.io/pkg/errors"
)

func StartValidationFlow() *ValidationFlow {
	flow := ValidationFlow{}
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

func (this *ValidationFlow) Step(fn func(vErrs *ft.ValidationErrors) error, endIfValidationError ...bool) (out *ValidationFlow) {
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

	if this.vErrs.Count() > 0 && len(endIfValidationError) > 0 && endIfValidationError[0] {
		this.skip = true
	}
	return
}

func (this *ValidationFlow) End() (*ft.ValidationErrors, error) {
	return this.vErrs, this.err
}
