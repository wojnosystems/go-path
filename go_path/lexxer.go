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

type nextState struct {
	triggeringToken token
	next            *lexerState
}

type lexerState struct {
	parse      func(l *lexer) *lexerState
	nextStates []nextState
}

var (
	stateEnd                    *lexerState
	stateError                  *lexerState
	stateItemVariableName       *lexerState
	stateItemSquareBracketOpen  *lexerState
	stateItemSquareBracketClose *lexerState
	stateItemMapKey             *lexerState
	stateItemArrayIndex         *lexerState
	stateItemDot                *lexerState
	stateItemMapKeyString       *lexerState
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
	stateEnd = &lexerState{
		nextStates: []nextState{},
	}
	stateError = &lexerState{
		nextStates: []nextState{},
	}

	stateItemVariableName = &lexerState{
		nextStates: make([]nextState, 0),
	}
	stateItemSquareBracketOpen = &lexerState{
		nextStates: make([]nextState, 0),
	}
	stateItemSquareBracketClose = &lexerState{
		nextStates: make([]nextState, 0),
	}
	stateItemMapKey = &lexerState{
		nextStates: make([]nextState, 0),
	}
	stateItemArrayIndex = &lexerState{
		nextStates: make([]nextState, 0),
	}
	stateItemDot = &lexerState{
		nextStates: make([]nextState, 0),
	}
	stateItemMapKeyString = &lexerState{
		nextStates: make([]nextState, 0),
	}

	stateStart = &lexerState{
		nextStates: make([]nextState, 0),
	}
}

func moveToNextState(states []nextState, t token) *lexerState {
	for _, state := range states {
		if state.triggeringToken == t {
			return state.next
		}
	}
	return stateError
}

func (l *lexer) returnStateError(err error) *lexerState {
	copy(l.currentValue, []rune(err.Error()))
	l.emit(itemError)
	return stateError
}

func isIdentRune(r rune) bool {
	return r < utf8.MaxRune && !unicode.IsSpace(r)
}

func (l *lexer) handleEOFOrError(err error, emitEvent token) *lexerState {
	if err != nil && err != io.EOF {
		return l.returnStateError(err)
	}
	if err == io.EOF {
		l.emit(emitEvent)
		return moveToNextState(stateStart.nextStates, itemEOF)
	}
	return nil
}

func (l *lexer) moveToTokenNextState(r rune, nextStates []nextState) *lexerState {
	if t, ok := key[string(r)]; ok {
		l.ignore() // consume the peeked value
		return moveToNextState(nextStates, t)
	}
	return nil
}

func linkNextStates() {
	stateStart.nextStates = append(stateStart.nextStates,
		nextState{
			triggeringToken: itemVariableName,
			next:            stateItemVariableName,
		},
		nextState{
			triggeringToken: itemSquareBracketOpen,
			next:            stateItemSquareBracketOpen,
		},
		nextState{
			triggeringToken: itemEOF,
			next:            stateEnd,
		},
	)
	stateStart.parse = func(l *lexer) *lexerState {
		r, err := l.peek()
		if nextState := l.handleEOFOrError(err, itemEOF); nextState != nil {
			return nextState
		}
		if nextState := l.moveToTokenNextState(r, stateStart.nextStates); nextState != nil {
			return nextState
		}
		if isIdentRune(r) {
			return stateItemVariableName
		} else {
			// invalid character
			return l.returnStateError(fmt.Errorf("unexpected rune: '%c'", r))
		}
	}

	stateEnd.parse = func(l *lexer) *lexerState {
		l.emit(itemEOF)
		return nil
	}

	stateItemVariableName.nextStates = append(stateItemVariableName.nextStates,
		nextState{
			triggeringToken: itemDot,
			next:            stateItemDot,
		},
		nextState{
			triggeringToken: itemSquareBracketOpen,
			next:            stateItemSquareBracketOpen,
		},
		nextState{
			triggeringToken: itemEOF,
			next:            stateEnd,
		},
	)
	stateItemVariableName.parse = func(l *lexer) *lexerState {
		for {
			r, err := l.peek()
			if nextState := l.handleEOFOrError(err, itemVariableName); nextState != nil {
				return nextState
			}

			if t, ok := key[string(r)]; ok {
				l.ignore() // consume the peeked value
				l.emit(itemVariableName)
				return moveToNextState(stateItemVariableName.nextStates, t)
			}
			if isIdentRune(r) {
				err = l.accept()
				if err != nil {
					return l.returnStateError(err)
				}
			} else {
				// invalid character
				return l.returnStateError(fmt.Errorf("unexpected rune: '%c'", r))
			}
		}
	}

	stateItemDot.nextStates = append(stateItemDot.nextStates,
		nextState{
			triggeringToken: itemVariableName,
			next:            stateItemVariableName,
		},
	)
	stateItemDot.parse = func(l *lexer) *lexerState {
		r, err := l.peek()
		if nextState := l.handleEOFOrError(err, itemError); nextState != nil {
			return nextState
		}
		if isIdentRune(r) {
			return stateItemVariableName
		} else {
			// invalid character
			return l.returnStateError(fmt.Errorf("unexpected rune: '%c'", r))
		}
	}

	stateItemSquareBracketOpen.nextStates = append(stateItemSquareBracketOpen.nextStates,
		nextState{
			triggeringToken: itemArrayIndex,
			next:            stateItemArrayIndex,
		},
		nextState{
			triggeringToken: itemQuoteDouble,
			next:            stateItemMapKeyString,
		},
	)
	stateItemSquareBracketOpen.parse = func(l *lexer) *lexerState {
		r, err := l.peek()
		if nextState := l.handleEOFOrError(err, itemError); nextState != nil {
			return nextState
		}
	}

	stateItemMapKeyString.nextStates = append(stateItemMapKeyString.nextStates,
		nextState{
			triggeringToken: itemMapKey,
			next:            stateItemMapKey,
		},
	)

	stateItemMapKey.nextStates = append(stateItemMapKey.nextStates,
		nextState{
			triggeringToken: itemQuoteDouble,
			next:            stateItemSquareBracketClose,
		},
	)

	stateItemArrayIndex.nextStates = append(stateItemArrayIndex.nextStates,
		nextState{
			triggeringToken: itemSquareBracketClose,
			next:            stateItemSquareBracketClose,
		},
	)

	stateItemSquareBracketClose.nextStates = append(stateItemSquareBracketClose.nextStates,
		nextState{
			triggeringToken: itemDot,
			next:            stateItemDot,
		},
		nextState{
			triggeringToken: itemSquareBracketOpen,
			next:            stateItemSquareBracketOpen,
		},
		nextState{
			triggeringToken: itemEOF,
			next:            stateEnd,
		},
	)
}

func init() {
	initializeStateVariables()
	linkNextStates()
}
