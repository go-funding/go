package errors

import . "errors"

func Wrap[Val any](prefix string) func(Val, error) (Val, error) {
	return func(v Val, err error) (Val, error) {
		if err != nil {
			return v, Join(New(prefix), err)
		}
		return v, err
	}
}
