package protoabs

import (
	"errors"
	"fmt"
	"github.com/oleiade/reflections"
	"strings"
)

var Opts *Options

type Options struct {
	MessageParentClass string
	ReservedPrefix     string
}

func ParseOptions(parameter *string) error {
	Opts = &Options{
		MessageParentClass: "Google\\Protobuf\\Internal\\Message",
		ReservedPrefix:     "PB",
	}

	if parameter != nil {
		args := strings.Split(*parameter, ",")
		for _, a := range args {
			parts := strings.Split(a, "=")
			opt := parts[0]
			val := parts[1]

			ok, err := reflections.HasField(Opts, opt)
			if err != nil {
				return err
			}
			if ok {
				err = reflections.SetField(Opts, opt, val)
				if err != nil {
					return err
				}
			} else {
				return errors.New(fmt.Sprintf("option %s does not exist", opt))
			}
		}
	}

	return nil
}
