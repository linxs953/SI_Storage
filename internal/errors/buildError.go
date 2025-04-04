package errors

import (
	"strings"

	pkgerr "github.com/pkg/errors"
)

// 自定义error的构建函数
// 自定义错误类型
type Error struct {
	Code     ErrorCode              // 错误代码
	Message  string                 // 可读的错误信息
	Details  map[string]interface{} // 扩展的上下文信息
	Internal error                  // 内部嵌套的错误
	// stack    []uintptr              // 调用堆栈
}

// 实现 error 接口
func (e *Error) Error() string {
	var parts []string
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if e.Internal != nil {
		parts = append(parts, "caused by: "+e.Internal.Error())
	}
	return strings.Join(parts, ", ")
}

// 创建新错误（自动捕获堆栈）
func New(code ErrorCode) *Error {
	return &Error{
		Code:     code,
		Message:  codeMessages[code],
		Internal: pkgerr.New(codeMessages[code]),
	}
}

func NewWithError(err error, code ErrorCode) *Error {
	return &Error{
		Code:     code,
		Message:  codeMessages[code],
		Internal: err,
	}
}

// 添加上下文信息
func (e *Error) WithDetails(key string, value interface{}) *Error {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value

	// 构建错误链路
	e.Internal = pkgerr.Wrap(e.Internal, key)

	// 每次更新details的时候，实时更新到message中，保证message是最新的
	e.Message = key
	return e
}

func (e *Error) UnWrap() error {
	if e == nil || e.Internal == nil {
		return nil
	}
	return pkgerr.Unwrap(e.Internal)
}

// 错误检查工具函数
func Is(err error, code ErrorCode) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}

// // 格式化输出（实现Formatter接口）
// func (e *Error) Format(s fmt.State, verb rune) {
// 	switch verb {
// 	case 'v':
// 		if s.Flag('+') {
// 			fmt.Fprintf(s, "[%d] %s\n", e.Code, e.Message)
// 			if len(e.Details) > 0 {
// 				fmt.Fprintln(s, "Details:")
// 				for k, v := range e.Details {
// 					fmt.Fprintf(s, "  %s: %+v\n", k, v)
// 				}
// 			}
// 			if e.Internal != nil {
// 				fmt.Fprint(s, "Caused by: ")
// 				if f, ok := e.Internal.(fmt.Formatter); ok {
// 					f.Format(s, verb)
// 				} else {
// 					fmt.Fprint(s, e.Internal.Error())
// 				}
// 			}
// 			printStack(s, e.stack)
// 			return
// 		}
// 		fallthrough
// 	case 's':
// 		fmt.Fprint(s, e.Error())
// 	case 'q':
// 		fmt.Fprintf(s, "%q", e.Error())
// 	}
// }

// 打印堆栈跟踪（实际实现需要更完善的格式）
// func printStack(s fmt.State, stack []uintptr) {
// 	fmt.Fprintln(s, "\nCall Stack:")
// 	frames := runtime.CallersFrames(stack)
// 	for {
// 		frame, more := frames.Next()
// 		fmt.Fprintf(s, "-> %s:%d %s\n", frame.File, frame.Line, frame.Function)
// 		if !more {
// 			break
// 		}
// 	}
// }

// 获取堆栈信息（实际使用时可以格式化为字符串）
// func (e *Error) StackTrace() []uintptr {
// 	return e.stack
// }

// 堆栈捕获（跳过指定层数）
// func captureStack(skip int) []uintptr {
// 	const depth = 32
// 	var pcs [depth]uintptr
// 	n := runtime.Callers(skip+1, pcs[:])
// 	return pcs[0:n]
// }

func (e *Error) GetMessage() string {
	if e == nil {
		return ""
	}
	return e.Message
}
