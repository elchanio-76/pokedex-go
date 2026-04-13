package input

import (
	"bufio"
	"fmt"
	"os"
)

// raw input handling with x/term

type LineReader interface {
    ReadLine(prompt string) (string, error)
    SetHistory([]string)
    Close() error
}
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

func (lr *lineReader) SetHistory(history []string) {
    // not supported with plain scanner
    
}

func (lr *lineReader) Close() error {
    // nothing to do
    return nil
}

func ReadLine() {
	
}
func NewLineReader() LineReader {
    return &lineReader{
        reader: bufio.NewScanner(os.Stdin),
    }
}
