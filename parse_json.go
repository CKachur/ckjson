package ckjson

import (
	"fmt"
	"strings"

	"github.com/CKachur/ckio"
)

func parseTrue(peekableRuneReader ckio.PeekableRuneReader) (string, error) {
	return parseRunes(peekableRuneReader, []rune{'t', 'r', 'u', 'e'})
}

func parseFalse(peekableRuneReader ckio.PeekableRuneReader) (string, error) {
	return parseRunes(peekableRuneReader, []rune{'f', 'a', 'l', 's', 'e'})
}

func parseNull(peekableRuneReader ckio.PeekableRuneReader) (string, error) {
	return parseRunes(peekableRuneReader, []rune{'n', 'u', 'l', 'l'})
}

func parseRunes(peekableRuneReader ckio.PeekableRuneReader, expectedRunes []rune) (string, error) {
	runeArray := make([]rune, len(expectedRunes))
	for i := 0; i < len(expectedRunes); i++ {
		runeValue, err := peekableRuneReader.Read()
		if err != nil {
			return string(runeArray[:i]), NewReadError(err.Error())
		}
		if runeValue != expectedRunes[i] {
			return string(runeArray[:i]), NewSyntaxError(fmt.Sprintf("expected '%c', found '%c'", expectedRunes[i], runeValue))
		}
		runeArray[i] = runeValue
	}
	return string(runeArray), nil
}

func parseString(peekableRuneReader ckio.PeekableRuneReader) (string, error) {
	runeValue, err := peekableRuneReader.Read()
	if err != nil {
		return "", NewReadError(err.Error())
	}
	if runeValue != '"' {
		return "", NewSyntaxError(fmt.Sprintf("expected '\"' to begin string, found '%c'", runeValue))
	}

	var stringBuilder strings.Builder
	stringBuilder.WriteRune(runeValue)

	for {
		runeValue, err = peekableRuneReader.Peek()
		if err != nil {
			return stringBuilder.String(), NewReadError(err.Error())
		}
		switch runeValue {
		case '"':
			peekableRuneReader.Read()
			stringBuilder.WriteRune(runeValue)
			return stringBuilder.String(), nil
		case '\\':
			returnedString, err := parseEscapeSequence(peekableRuneReader)
			stringBuilder.WriteString(returnedString)
			if err != nil {
				return stringBuilder.String(), err
			}
		default:
			if isControlCharacter(runeValue) {
				return stringBuilder.String(), NewSyntaxError(fmt.Sprintf("unexpected control character 0x%X", runeValue))
			}
			peekableRuneReader.Read()
			stringBuilder.WriteRune(runeValue)
		}
	}
}

func isControlCharacter(runeValue rune) bool {
	if runeValue >= 0 && runeValue <= 0x1F {
		return true
	}
	return false
}

func parseEscapeSequence(peekableRuneReader ckio.PeekableRuneReader) (string, error) {
	runeValue, err := peekableRuneReader.Read()
	if err != nil {
		return "", NewReadError(err.Error())
	}
	if runeValue != '\\' {
		return "", NewSyntaxError(fmt.Sprintf("expected '\\' at beginning of escape sequence, found '%c'", runeValue))
	}

	runeValue, err = peekableRuneReader.Read()
	if err != nil {
		return "\\", NewReadError(err.Error())
	}
	switch runeValue {
	case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
		return fmt.Sprintf("\\%c", runeValue), nil
	case 'u':
		escapeHexCodePointSting, err := parseHexadecimalCodePoint(peekableRuneReader)
		return fmt.Sprintf("\\u%s", escapeHexCodePointSting), err
	default:
		return "\\", NewSyntaxError(fmt.Sprintf("invalid escape sequence character '%c'", runeValue))
	}
}

func parseHexadecimalCodePoint(peekableRuneReader ckio.PeekableRuneReader) (string, error) {
	var runeArray []rune = []rune{0, 0, 0, 0}
	for i := 0; i < 4; i++ {
		runeValue, err := peekableRuneReader.Read()
		if err != nil {
			return string(runeArray[:i]), NewReadError(err.Error())
		}
		if !isHexadecimalCharacter(runeValue) {
			return string(runeArray[:i]), NewSyntaxError(fmt.Sprintf("invalid hexadecimal character in code point '%c'", runeValue))
		}
		runeArray[i] = runeValue
	}
	return string(runeArray), nil
}

func isHexadecimalCharacter(runeValue rune) bool {
	switch runeValue {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'A', 'B', 'C', 'D', 'E', 'F':
		return true
	default:
		return false
	}
}
