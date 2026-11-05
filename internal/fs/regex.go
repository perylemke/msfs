package fs

import "regexp"

func regexpCompile(re string) (matcher, error) {
	r, err := regexp.Compile(re)
	if err != nil {
		return nil, err
	}
	return r, nil
}
