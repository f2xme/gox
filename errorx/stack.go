package errorx

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// StackFrame 表示堆栈跟踪中的单个帧
type StackFrame struct {
	// File 源文件名
	File string
	// Line 行号
	Line int
	// Function 函数名
	Function string
}

func (f StackFrame) String() string {
	return fmt.Sprintf("%s:%d %s", filepath.Base(f.File), f.Line, f.Function)
}

// captureStack 捕获当前的堆栈跟踪
// skip 是要跳过的堆栈帧数量
func captureStack(skip int) []StackFrame {
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(skip, pcs[:])

	frames := make([]StackFrame, 0, n)
	for i := range n {
		pc := pcs[i]
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line := fn.FileLine(pc)
		frames = append(frames, StackFrame{
			File:     file,
			Line:     line,
			Function: fn.Name(),
		})
	}
	return frames
}
