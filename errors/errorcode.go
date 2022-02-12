package errors

type BizCode struct {
	BizCode    int           `json:"http_code" yaml:"http_code"`
	BizMessage string        `json:"message" yaml:"message"`
	BizDetail  BizCodeDetail `json:"detail" yaml:"detail"`
}
type BizCodeDetail struct {
	Code  string `json:"code" yaml:"code"`
	Error string `json:"error" yaml:"error"`
}


func (c BizCode) Code() int {
	return c.BizCode
}

func (c BizCode) Message() string {
	return c.BizMessage
}

func (c BizCode) Detail() interface{} {
	return c.BizDetail
}

func (c *BizCode) SetErrorContent(error string) *BizCode{
	c.BizDetail.Error = error
	return c
}

func newCode(errCode string, httpCode int, message string) BizCode {
	return BizCode{
		BizCode:    httpCode,
		BizMessage: message,
		BizDetail: BizCodeDetail{
			Code:  errCode,
			Error: "",
		},
	}
}

var (
	OK                   = newCode("e.20000", 200, "OK")
	NotFound             = newCode("e.40001", 200, "Not Found")
	NotConfigured        = newCode("e.40002", 200, "Not Configured")
	InternalError        = newCode("e.40003", 500, "Internal Error")
	SQLError             = newCode("e.40004", 500, "Sql Execute Error")
	StructConvertError   = newCode("e.40005", 500, "Convert To Specify Struct Error")
	DataTypeConvertError = newCode("e.40006", 500, "Convert To Specify Struct Error")
	WrongParameterError  = newCode("e.40007", 500, "Parameter Wrong Error")
)
