package protoabs

import (
	"errors"
	"fmt"
	"github.com/oleiade/reflections"
	"strings"
)

type Options struct {
	MessageParentClass string
}

func ParseOptions(parameter *string) (*Options, error) {
	o := &Options{
		MessageParentClass: "Google\\Protobuf\\Internal\\Message",
	}

	if parameter != nil {
		args := strings.Split(*parameter, ",")
		for _, a := range args {
			parts := strings.Split(a, "=")
			opt := parts[0]
			val := parts[1]

			ok, err := reflections.HasField(o, opt)
			if err != nil {
				return nil, err
			}
			if ok {
				err = reflections.SetField(o, opt, val)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.New(fmt.Sprintf("option %s does not exist", opt))
			}
		}
	}

	return o, nil
}
