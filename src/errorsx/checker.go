package errorsx

type Check func(err error) error

func Is(err error, check Check) bool {
	if check == nil {
		return false
	}
	return check(err) != nil
}

func Convert(err error, check Check) error {
	if check == nil {
		return err
	}

	if newErr := check(err); newErr != nil {
		return newErr
	}
	return err
}

func Or(checks ...Check) Check {
	return func(err error) error {
		if err == nil {
			return nil
		}

		for _, convertor := range checks {
			if err := convertor(err); err != nil {
				return err
			}
		}

		return nil
	}
}

func And(checks ...Check) Check {
	return func(err error) error {
		var lastResponse error
		for _, check := range checks {
			lastResponse = check(err)
			if lastResponse == nil {
				return nil
			}
		}
		return lastResponse
	}
}
