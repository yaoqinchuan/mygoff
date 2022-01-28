package utils

import (
	"fmt"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gregex"
)

type MyLoggerWriter struct {
	logger *glog.Logger
}

func (w *MyLoggerWriter) Write(p []byte) (n int, err error) {
	var (
		s = string(p)
	)
	if gregex.IsMatchString(`PANI|FATA`, s) {
		fmt.Println("SERIOUS ISSUE OCCURRED!! I'd better tell monitor in first time!")
		return 0, nil
	}
	return w.logger.Write(p)
}
