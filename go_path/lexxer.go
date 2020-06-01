package go_path

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/wojnosystems/go-optional"
	"io"
	"unicode"
	"unicode/utf8"
)

// Parses go-path strings into the Pather object that it represents
// Grammar:
// There are 3 object types (so far): Structs, Arrays (slices), and Maps.
// Roots. Each of the above type may be a root, in which case, the string looks like:
// * nameOfVarInStruct
// * [indexOfArray]
// * ["keyOfMap"]
// Struct roots do not have the leading dot, but the dot separates structs from each other:
// * nameOfVar.anotherVar.yetAnotherVar to indicate a nested struct like:
// type Third struct {
//   yetAnotherVar
// }
// type Second struct {
//   anotherVar Third
// }
// type Root struct {
//   nameOfVar Second
// }
// This is a very simple lexxer as the grammar does not support nested square brackets or anything else that's nested within itself, so keeping a paren level is not necessary
//
// newlines are ignored at this time

type token int

const (
	itemError token = iota
	itemEOF

	literalBegin
	itemArrayIndex // 12345
	itemMapKey     // key
	literalEnd

	itemVariableName // variableName

	symbolBegin
	itemDot                // .
	itemQuoteDouble        // "
	itemSquareBracketOpen  // [
	itemSquareBracketClose // ]
	symbolEnd
)

type item struct {
	typ  token
	val  string
	line uint
	col  uint
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ == itemVariableName:
		return fmt.Sprintf("<%s>", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

const (
	spaceChars = " \t\r\n"
	IdentMax   = 1024
)

var key = map[string]token{
	"[":  itemSquareBracketOpen,
	"]":  itemSquareBracketClose,
	".":  itemDot,
	"\"": itemQuoteDouble,
}

type lexerState struct {
	parse func(l *lexer) *lexerState
}

var (
	stateEnd                    *lexerState
	stateError                  *lexerState
	stateItemVariableName       *lexerState
	stateItemSquareBracketOpen  *lexerState
	stateItemSquareBracketClose *lexerState
	stateItemMapKey             *lexerState
	stateItemMapKeyEnd          *lexerState
	stateItemArrayIndex         *lexerState
	stateItemDot                *lexerState
	stateStart                  *lexerState
)

type lexer struct {
	source       *bufio.Reader
	peeked       optional.Rune
	currentState *lexerState
	line         uint // 1 + number of newlines seen
	col          uint
	startCol     uint
	itemEmitter  chan item
	currentValue []rune
	startLine    uint
}

func newLexer(source io.Reader) lexer {
	return lexer{
		source:       bufio.NewReader(source),
		currentState: stateStart,
		itemEmitter:  make(chan item),
		line:         1,
		col:          1,
		startLine:    1,
		startCol:     1,
		currentValue: make([]rune, 0, IdentMax),
	}
}

func (l *lexer) appendCurrent(r rune) error {
	if len(l.currentValue) == cap(l.currentValue) {
		return errors.New("identifier longer than our maximum buffer size")
	}
	l.currentValue = append(l.currentValue, r)
	return nil
}

func (l *lexer) lex() {
	for state := stateStart; state != nil; {
		state = state.parse(l)
	}
	close(l.itemEmitter)
}

func (l *lexer) peek() (r rune, err error) {
	if l.peeked.IsSet() {
		r = l.peeked.Value()
		return
	}
	r, _, err = l.source.ReadRune()
	if err != nil {
		return
	}
	l.peeked.Set(r)
	return
}

// ignore the last rune appended
func (l *lexer) ignore() {
	if l.peeked.IsSet() {
		l.peeked.Unset()
		l.col++
	}
}

func (l *lexer) accept() (err error) {
	var r rune
	if l.peeked.IsSet() {
		r = l.peeked.Value()
		l.peeked.Unset()
		err = l.appendCurrent(r)
		if err != nil {
			return
		}
	}
	l.col++
	if r == '\n' {
		l.line++
		l.col = 1
	}
	return
}

func (l *lexer) emit(t token) {
	l.itemEmitter <- item{typ: t, col: l.startCol, val: string(l.currentValue), line: l.startLine}
	l.startLine = l.line
	l.currentValue = l.currentValue[0:0]
	l.startCol = l.col
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) getNextItem() item {
	return <-l.itemEmitter
}

// drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) drain() {
	for range l.itemEmitter {

	}
}

func initializeStateVariables() {
	stateEnd = &lexerState{}
	stateError = &lexerState{}

	stateItemVariableName = &lexerState{}
	stateItemSquareBracketOpen = &lexerState{}
	stateItemSquareBracketClose = &lexerState{}
	stateItemMapKey = &lexerState{}
	stateItemMapKeyEnd = &lexerState{}
	stateItemArrayIndex = &lexerState{}
	stateItemDot = &lexerState{}

	stateStart = &lexerState{}
}

func (l *lexer) returnStateError(err error) *lexerState {
	copy(l.currentValue, []rune(err.Error()))
	l.emit(itemError)
	return stateError
}

func isIdentRune(r rune) bool {
	return r < utf8.MaxRune && !unicode.IsSpace(r) && r != '.'
}

func isNumber(r rune) bool {
	return '0' <= r && r <= '9'
}

func isQuoteDouble(r rune) bool {
	return '"' == r
}

func isEscapeChar(r rune) bool {
	return '\\' == r
}

func (l *lexer) handleEOFOrError(err error, emitEventIfEOF token) *lexerState {
	if err != nil && err != io.EOF {
		return l.returnStateError(err)
	}
	if err == io.EOF {
		l.emit(emitEventIfEOF)
		return stateEnd
	}
	return nil
}

func (l *lexer) returnErrorUnexpectedRune(r rune) *lexerState {
	return l.returnStateError(fmt.Errorf("unexpected rune: '%c'", r))
}

func linkNextStates() {
	stateStart.parse = func(l *lexer) *lexerState {
		r, err := l.peek()
		if nextState := l.handleEOFOrError(err, itemEOF); nextState != nil {
			return nextState
		}
		switch {
		case '[' == r:
			l.ignore()
			return stateItemSquareBracketOpen
		case isIdentRune(r):
			return stateItemVariableName
		default:
			return l.returnErrorUnexpectedRune(r)
		}
	}

	stateEnd.parse = func(l *lexer) *lexerState {
		l.emit(itemEOF)
		return nil
	}

	stateItemVariableName.parse = func(l *lexer) *lexerState {
		for {
			r, err := l.peek()
			if nextState := l.handleEOFOrError(err, itemVariableName); nextState != nil {
				return nextState
			}
			switch {
			case '.' == r:
				l.ignore()
				l.emit(itemVariableName)
				return stateItemDot
			case '[' == r:
				l.ignore()
				l.emit(itemVariableName)
				return stateItemSquareBracketOpen
			case isIdentRune(r):
				err = l.accept()
				if err != nil {
					return l.returnStateError(err)
				}
			default:
				return l.returnErrorUnexpectedRune(r)
			}
		}
	}

	stateItemDot.parse = func(l *lexer) *lexerState {
		r, err := l.peek()
		if nextState := l.handleEOFOrError(err, itemError); nextState != nil {
			return nextState
		}
		if isIdentRune(r) {
			return stateItemVariableName
		} else {
			// invalid character
			return l.returnErrorUnexpectedRune(r)
		}
	}

	stateItemSquareBracketOpen.parse = func(l *lexer) *lexerState {
		r, err := l.peek()
		if nextState := l.handleEOFOrError(err, itemError); nextState != nil {
			return nextState
		}
		switch {
		case isQuoteDouble(r):
			l.ignore()
			return stateItemMapKey
		case isNumber(r):
			return stateItemArrayIndex
		default:
			return l.returnErrorUnexpectedRune(r)
		}
	}

	stateItemMapKey.parse = func(l *lexer) *lexerState {
		isEscaped := false
		for {
			r, err := l.peek()
			if nextState := l.handleEOFOrError(err, itemError); nextState != nil {
				return nextState
			}
			if !isEscaped && isQuoteDouble(r) {
				l.ignore()
				l.emit(itemMapKey)
				return stateItemMapKeyEnd
			}
			if isEscapeChar(r) {
				if isEscaped {
					isEscaped = false
				} else {
					isEscaped = true
				}
			}
			err = l.accept()
			if nil != err {
				return l.returnStateError(err)
			}
		}
	}

	stateItemMapKeyEnd.parse = func(l *lexer) *lexerState {
		r, err := l.peek()
		if nextState := l.handleEOFOrError(err, itemError); nextState != nil {
			return nextState
		}
		switch {
		case r != ']':
			return l.returnErrorUnexpectedRune(r)
		default:
			l.ignore()
			return stateItemSquareBracketClose
		}
	}

	stateItemArrayIndex.parse = func(l *lexer) *lexerState {
		for {
			r, err := l.peek()
			if nextState := l.handleEOFOrError(err, itemError); nextState != nil {
				return nextState
			}
			switch {
			case r == ']':
				l.ignore()
				l.emit(itemArrayIndex)
				return stateItemSquareBracketClose
			case isNumber(r):
				err = l.accept()
				if err != nil {
					return l.returnStateError(err)
				}
			default:
				return l.returnErrorUnexpectedRune(r)
			}
		}
	}

	stateItemSquareBracketClose.parse = func(l *lexer) *lexerState {
		r, err := l.peek()
		if nextState := l.handleEOFOrError(err, itemEOF); nextState != nil {
			return nextState
		}
		switch r {
		case '.':
			l.ignore()
			return stateItemDot
		case '[':
			l.ignore()
			return stateItemSquareBracketOpen
		default:
			return l.returnErrorUnexpectedRune(r)
		}
	}
}

func init() {
	initializeStateVariables()
	linkNextStates()
}
