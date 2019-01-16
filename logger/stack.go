package logger

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

const maxStackSize = 32

// Frame identifies a file, line & function name in the stack.
type frame struct {
	File string
	Line int
	Name string
}

// String provides the standard file:line representation.
func (f frame) String() string {
	return fmt.Sprintf("%s:%d %s", f.File, f.Line, f.Name)
}

// Stack represents an ordered set of Frames.
type stack []frame

// String provides the standard multi-line stack trace.
func (s stack) String() string {
	var b bytes.Buffer
	writeStack(&b, s)
	return b.String()
}

// Callers returns a Stack of Frames for the callers. The argument skip is the
// number of stack frames to ascend, with 0 identifying the caller of Callers.
func callers(skip int) stack {
	pcs := make([]uintptr, maxStackSize)
	num := runtime.Callers(skip+2, pcs)
	stack := make(stack, num)
	for i, pc := range pcs[:num] {
		fun := runtime.FuncForPC(pc)
		file, line := fun.FileLine(pc - 1)
		stack[i].File = stripGOPATH(file)
		stack[i].Line = line
		stack[i].Name = stripPackage(fun.Name())
	}
	return stack
}

func writeStack(b *bytes.Buffer, s stack) {
	var width int
	for _, f := range s {
		if l := len(f.File) + numDigits(f.Line) + 1; l > width {
			width = l
		}
	}
	b.WriteRune('\n')
	for _, f := range s {
		b.WriteString(f.File)
		b.WriteRune(rune(':'))
		n, _ := fmt.Fprintf(b, "%d", f.Line)
		for i := width - len(f.File) - n; i != 0; i-- {
			b.WriteRune(rune(' '))
		}
		b.WriteString(f.Name)
		b.WriteRune(rune('\n'))
	}
}

func numDigits(i int) int {
	var n int
	for {
		n++
		i = i / 10
		if i == 0 {
			return n
		}
	}
}

func stripGOPATH(f string) string {
	if i := strings.Index(f, "/src/"); i >= 0 {
		return f[i+5:]
	} else {
		return f
	}
}

// StripPackage strips the package name from the given Func.Name.
func stripPackage(n string) string {
	slashI := strings.LastIndex(n, "/")
	if slashI == -1 {
		slashI = 0 // for built-in packages
	}
	dotI := strings.Index(n[slashI:], ".")
	if dotI == -1 {
		return n
	}
	return n[slashI+dotI+1:]
}
