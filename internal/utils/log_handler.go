package utils

import (
	"context"
	"encoding/json"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
	"os"
)

type JsonOutPutsForLogger struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Content string `json:"content"`
}

//glog组件提供了超级强大的、可自定义日志处理的Handler特性。Handler采用了中间件设计方式，开发者可以为日志对象注册多个处理Handler，
//也可以在Handler中覆盖默认的日志组件处理逻辑。可以看到第二个参数为日志处理的日志信息，并且为指针类型，
//意味着在Handler中可以修改该参数的任意属性信息，并且修改后的内容将会传递给下一个Handler。这样就可以将内容输出到Es或者群里
var loggingJsonHandler glog.Handler = func(ctx context.Context, in *glog.HandlerInput) {
	jsonForLogger := JsonOutPutsForLogger{
		in.TimeFormat,
		gstr.Trim(in.LevelFormat, "[]"),
		gstr.Trim(in.Content),
	}
	jsonBytes, err := json.Marshal(jsonForLogger)
	if err != nil {
		_, _ = os.Stdout.WriteString(err.Error())
		return
	}
	in.Buffer.Write(jsonBytes)
	in.Buffer.WriteString("\n")
	in.Next()
}
