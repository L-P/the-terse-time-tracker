//+build !fixture

package tt

import "errors"

func (tt *TT) Fixture() error {
	return errors.New("not implemented") // nolint:goerr113
}
