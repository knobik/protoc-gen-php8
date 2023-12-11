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

func ParseOptions(parameter string) (*Options, error) {
	o := &Options{
		MessageParentClass: "Google\\Protobuf\\Internal\\Message",
	}

	args := strings.Split(parameter, ",")
	for _, a := range args {
		parts := strings.Split(a, "=")

		ok, err := reflections.HasField(o, parts[0])
		if err != nil {
			return nil, err
		}
		if ok {
			err = reflections.SetField(o, parts[0], parts[1])
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New(fmt.Sprintf("option %s does not exist", parts[0]))
		}
	}

	return o, nil
}
