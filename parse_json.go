package ckjson

import (
	"fmt"
	"strings"

	"github.com/CKachur/ckio"
)

func parseTrue(peekableRuneReader ckio.PeekableRuneReader) (string, error) {
	return parseExactRunes(peekableRuneReader, []rune{'t', 'r', 'u', 'e'})
}

func parseFalse(peekableRuneReader ckio.PeekableRuneReader) (string, error) {
	return parseExactRunes(peekableRuneReader, []rune{'f', 'a', 'l', 's', 'e'})
}

func parseNull(peekableRuneReader ckio.PeekableRuneReader) (string, error) {
	return parseExactRunes(peekableRuneReader, []rune{'n', 'u', 'l', 'l'})
}

func parseExactRunes(peekableRuneReader ckio.PeekableRuneReader, expectedRunes []rune) (string, error) {
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

func parseNumber(peekableRuneReader ckio.PeekableRuneReader) (string, error) {
	var stringBuilder strings.Builder

	minusString, err := parseSingleRuneIfEqual(peekableRuneReader, '-')
	if err != nil {
		return "", err
	}
	stringBuilder.WriteString(minusString)

	preDecimalDigitsString, err := parsePreDecimalDigits(peekableRuneReader)
	stringBuilder.WriteString(preDecimalDigitsString)
	if err != nil {
		return stringBuilder.String(), err
	}

	runeValue, peekErr := peekableRuneReader.Peek()
	shouldCheckPeekErrorAgain := false
	if peekErr != nil {
		return stringBuilder.String(), nil
	}
	if runeValue == '.' {
		decimalString, err := parseSingleRuneIfEqual(peekableRuneReader, '.')
		stringBuilder.WriteString(decimalString)
		if err != nil {
			return stringBuilder.String(), err
		}

		digitString := parseMultipleRunesIfInRange(peekableRuneReader, '0', '9')
		if digitString == "" {
			return stringBuilder.String(), NewSyntaxError("expected digit in number after decimal")
		}
		stringBuilder.WriteString(digitString)

		runeValue, peekErr = peekableRuneReader.Peek()
		shouldCheckPeekErrorAgain = true
	}

	if shouldCheckPeekErrorAgain && peekErr != nil {
		return stringBuilder.String(), nil
	}
	if runeValue == 'e' || runeValue == 'E' {
		exponentString, err := parseExponent(peekableRuneReader)
		stringBuilder.WriteString(exponentString)
		if err != nil {
			return stringBuilder.String(), err
		}
	}

	return stringBuilder.String(), nil
}

func parseExponent(peekableRuneReader ckio.PeekableRuneReader) (string, error) {
	var stringBuilder strings.Builder

	runeValue, err := peekableRuneReader.Peek()
	if err != nil {
		return "", NewReadError(err.Error())
	}
	if runeValue != 'e' && runeValue != 'E' {
		return "", NewSyntaxError(fmt.Sprintf("expected 'e' or 'E', found '%c'", runeValue))
	}
	peekableRuneReader.Read()
	stringBuilder.WriteRune(runeValue)

	runeValue, err = peekableRuneReader.Peek()
	if err != nil {
		return stringBuilder.String(), NewReadError(err.Error())
	}
	hasSign := false
	if runeValue == '+' || runeValue == '-' {
		hasSign = true
		peekableRuneReader.Read()
		stringBuilder.WriteRune(runeValue)
	}

	digitString := parseMultipleRunesIfInRange(peekableRuneReader, '0', '9')
	if digitString == "" {
		if !hasSign {
			return stringBuilder.String(), NewSyntaxError("expected '+', '-', or digit in exponent")
		}
		return stringBuilder.String(), NewSyntaxError("expected digit in exponent")
	}
	stringBuilder.WriteString(digitString)

	return stringBuilder.String(), nil
}

func parsePreDecimalDigits(peekableRuneReader ckio.PeekableRuneReader) (string, error) {
	var stringBuilder strings.Builder

	runeValue, err := peekableRuneReader.Peek()
	if err != nil {
		return "", NewReadError(err.Error())
	}
	switch runeValue {
	case '0':
		peekableRuneReader.Read()
		stringBuilder.WriteRune(runeValue)
		return stringBuilder.String(), nil
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		digitString := parseMultipleRunesIfInRange(peekableRuneReader, '0', '9')
		stringBuilder.WriteString(digitString)
		return stringBuilder.String(), nil
	default:
		return stringBuilder.String(), NewSyntaxError(fmt.Sprintf("expected digit, found '%c'", runeValue))
	}
}

func parseSingleRuneIfEqual(peekableRuneReader ckio.PeekableRuneReader, expectedRune rune) (string, error) {
	runeValue, err := peekableRuneReader.Peek()
	if err != nil {
		return "", NewReadError(err.Error())
	}
	if runeValue == expectedRune {
		peekableRuneReader.Read()
		return string(runeValue), nil
	}
	return "", nil
}

func parseMultipleRunesIfInRange(peekableRuneReader ckio.PeekableRuneReader, runeRangeStart, runeRangeEnd rune) string {
	var stringBuilder strings.Builder
	for {
		runeValue, err := peekableRuneReader.Peek()
		if err != nil {
			return stringBuilder.String()
		}
		if runeRangeStart <= runeValue && runeValue <= runeRangeEnd {
			peekableRuneReader.Read()
			stringBuilder.WriteRune(runeValue)
		} else {
			break
		}
	}
	return stringBuilder.String()
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
	if 0 <= runeValue && runeValue <= 0x1F {
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
	return ('0' <= runeValue && runeValue <= '9') ||
		('a' <= runeValue && runeValue <= 'f') ||
		('A' <= runeValue && runeValue <= 'F')
}
