package dynamicmodel

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

func StartValidationFlow(startWith ...Validatable) *ValidationFlow {
	flow := ValidationFlow{}
	if len(startWith) > 0 {
		return flow.Start().Step(func(cErrs *ft.ClientErrors) error {
			result := startWith[0].Validate()
			if result != nil {
				*cErrs = result
			}
			return nil
		})
	}
	return flow.Start()
}

func StartValidationFlowCopy(initClientErrs *ft.ClientErrors, startWith ...Validatable) *ValidationFlow {
	flow := ValidationFlow{}
	if len(startWith) > 0 {
		return flow.StartCopy(initClientErrs).Step(func(cErrs *ft.ClientErrors) error {
			result := startWith[0].Validate()
			if result != nil {
				*cErrs = result
			}
			return nil
		})
	}
	return flow.Start()
}

type ValidationFlow struct {
	cErrs *ft.ClientErrors
	err   error
	skip  bool
}

func (this *ValidationFlow) Start() *ValidationFlow {
	this.cErrs = ft.NewClientErrors()
	return this
}

func (this *ValidationFlow) StartCopy(initialErrs *ft.ClientErrors) *ValidationFlow {
	this.cErrs = ft.NewClientErrors()
	this.cErrs.Concat(*initialErrs)
	return this.Step(func(cErrs *ft.ClientErrors) error {
		// Stop flow immediately if cErrs has errors
		return nil
	})
}

func (this *ValidationFlow) StepS(fn func(vErrs *ft.ClientErrors, stop func()) error, ignoreValidationError ...bool) (out *ValidationFlow) {
	defer func() {
		out = this
		if e := recover(); e != nil {
			err, ok := e.(error)
			if ok {
				this.err = err
				return
			}
			this.err = errors.Errorf("ValidationFlow.Step: %v", e)
		}
	}()

	if this.err != nil || this.skip {
		return
	}

	this.err = fn(this.cErrs, func() {
		this.skip = true
	})

	if this.cErrs.Count() > 0 && (len(ignoreValidationError) == 0 || !ignoreValidationError[0]) {
		this.skip = true
	}
	return
}

func (this *ValidationFlow) Step(fn func(vErrs *ft.ClientErrors) error, ignoreValidationError ...bool) (out *ValidationFlow) {
	return this.StepS(func(vErrs *ft.ClientErrors, stop func()) error {
		return fn(vErrs)
	}, ignoreValidationError...)
}

func (this *ValidationFlow) End() (ft.ClientErrors, error) {
	return *this.cErrs, this.err
}

type Validatable interface {
	Validate() ft.ClientErrors
}
