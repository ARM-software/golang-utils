package step

func (o *StepOptions[I, O]) GetAction() ErrorAction {
	if o == nil {
		return ErrorActionStop
	}
	return o.action
}

func (o *StepOptions[I, O]) GetCompensatePrevious() bool {
	return o != nil && o.compensatePrevious
}

func (o *StepOptions[I, O]) GetCompensation() CompensationFunc[O] {
	if o == nil {
		return nil
	}
	return o.compensation
}

func (o *StepOptions[I, O]) GetFallback() FallbackFunc[I, O] {
	if o == nil {
		return nil
	}
	return o.fallback
}
