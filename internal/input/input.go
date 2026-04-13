package input

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

type Terminal interface {
	ReadLine(prompt string) (string, error)
	SetHistory([]string)
	Close() error
	Print(format string, args ...any)
	Println(args ...any)
}

// bufio.Scanner buffered input handling (simple, fall back)

type lineReader struct {
    reader *bufio.Scanner
}

func  (lr *lineReader) ReadLine(prompt string) (string, error) {
    fmt.Print(prompt)
    if !lr.reader.Scan() {
        return "", lr.reader.Err()
    }
		if err := lr.reader.Err(); err!=nil {
			return "", err
		}
    return lr.reader.Text(), nil
}

func (lr *lineReader) Print(format string, args ...any) {
	fmt.Printf(format, args...)
}

func (lr *lineReader) Println(args ...any) {
	fmt.Println(args...)
}

func (lr *lineReader) SetHistory(history []string) {
    // not supported with plain scanner
    
}

func (lr *lineReader) Close() error {
    // nothing to do
    return nil
}

func NewLineReader() Terminal {
	return &lineReader{
		reader: bufio.NewScanner(os.Stdin),
	}
}

// raw input handling with x/term

type termReader struct {
	term *term.Terminal
	oldState *term.State
	fd int
}

func NewTerminalReader() Terminal {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
					panic(err)
	}

	return &termReader{
		term: term.NewTerminal(os.Stdin, ""),
		oldState: oldState,
		fd: fd,
	}
}

func (tr *termReader) Close() error {
     
    return term.Restore(tr.fd, tr.oldState)
}

func (tr *termReader) SetHistory(history []string) {
	// to be implemented
}

func translateCRLF(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", "\r\n")
	return s
}

func (r *termReader) Print(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	r.term.Write([]byte(translateCRLF(s)))
}

func (tr *termReader) Println(args ...any) {
	s := fmt.Sprintln(args...)
	tr.term.Write([]byte(translateCRLF(s)))
}

func (tr *termReader) ReadLine(prompt string) (string, error) {
	// read each character 
	
	//tr.term.SetPrompt(prompt)
    
	var line []byte
	tr.term.Write([]byte(prompt))

	for {
			
			b, err := tr.readByte()
			if err != nil {
					return string(line), err
			}
			
			switch b {
			case '\r', '\n':
					tr.term.Write([]byte("\r\n"))
					return string(line), nil
			case 127, 8: // backspace
					if len(line) > 0 {
							line = line[:len(line)-1]
							tr.term.Write([]byte("\b \b")) // erase char
					}
			case 0x1b: // escape sequence (arrow keys)
					seq := make([]byte, 2)
					os.Stdin.Read(seq)
					if seq[0] == '[' {
							switch seq[1] {
							case 'A': // up arrow
									tr.term.Write([]byte("UP ARROW\r\n"))
							case 'B': // down arrow
									tr.term.Write([]byte("DOWN ARROW\r\n"))
							}
					}
			default:
					if len(line) >= 256 { // limit line length
							continue
					}
					line = append(line, b)
					tr.term.Write([]byte{b}) // echo
			}
	}
}


func (tr *termReader) readByte() (byte, error) {
    buf := make([]byte, 1)
    _, err := os.Stdin.Read(buf)
    return buf[0], err
}

