package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"mygogf/internal/cmd"
	_ "mygogf/internal/packed"
)

func main() {

	// 选项输入参数其实是一个map类型。其中键值为选项名称，
	//同一个选项的不同名称可以通过,符号进行分隔。比如，该示例中n和name选项是同一个选项，
	//当用户输入-n john的时候，n和name选项都会获得到数据john。
	//而键值是一个布尔类型，标识该选项是否需要解析参数。这一选项配置是非常重要的，因为有的选项是不需要获得数据的，
	//仅仅作为一个标识。例如，-f force这个输入，在需要解析数据的情况下，选项f的值为force；而在不需要解析选项数据的情况下，
	//其中的force便是命令行的一个参数，而不是选项。
	paras, err := gcmd.Parse(g.MapStrBool{
		"version, v":true,
	})
	if err != nil {
		g.Log("error occurs")
	}
	cmd.Main.Run(gctx.New())
}
