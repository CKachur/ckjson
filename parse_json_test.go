package ckjson

import (
	"bufio"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/CKachur/ckio"
)

const MIN_UNICODE_RUNE rune = 0
const MAX_UNICODE_RUNE rune = 0x10FFFF
const READ_RUNE_ERROR_EOF string = "could not read rune: EOF"

var VALID_HEX_CHARACTERS []rune = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'A', 'B', 'C', 'D', 'E', 'F'}

func TestParseTrueReturnsValidTrue(t *testing.T) {
	testIndividualParseFunctionCall(t, "true", "true", "", parseTrue)
}

func TestParseFalseReturnsValidFalse(t *testing.T) {
	testIndividualParseFunctionCall(t, "false", "false", "", parseFalse)
}

func TestParseNullReturnsValidNull(t *testing.T) {
	testIndividualParseFunctionCall(t, "null", "null", "", parseNull)
}

func TestParseEscapeSequenceReturnsValidEscapeSequence(t *testing.T) {
	testIndividualParseFunctionCall(t, "\\\"", "\\\"", "", parseEscapeSequence)
	testIndividualParseFunctionCall(t, "\\\\", "\\\\", "", parseEscapeSequence)
	testIndividualParseFunctionCall(t, "\\/", "\\/", "", parseEscapeSequence)
	testIndividualParseFunctionCall(t, "\\b", "\\b", "", parseEscapeSequence)
	testIndividualParseFunctionCall(t, "\\f", "\\f", "", parseEscapeSequence)
	testIndividualParseFunctionCall(t, "\\n", "\\n", "", parseEscapeSequence)
	testIndividualParseFunctionCall(t, "\\r", "\\r", "", parseEscapeSequence)
	testIndividualParseFunctionCall(t, "\\t", "\\t", "", parseEscapeSequence)

	numberOfValidHexCharacters := float64(len(VALID_HEX_CHARACTERS))
	for i := float64(0); i < math.Pow(numberOfValidHexCharacters, 4); i++ {
		inputString := fmt.Sprintf("\\u%s", getFourCharacterHexStringById(int(i)))
		testIndividualParseFunctionCall(t, inputString, inputString, "", parseEscapeSequence)
	}
}

func TestParseEscapeSequenceReturnsErrorForNonReverseSolidusStart(t *testing.T) {
	for runeValue := rune(MIN_UNICODE_RUNE); runeValue <= MAX_UNICODE_RUNE; runeValue++ {
		if runeValue != '\\' {
			inputString := string(runeValue)
			expectedOutputString := ""
			expectedErrorString := fmt.Sprintf("expected '\\' at beginning of escape sequence, found '%c'", runeValue)
			testIndividualParseFunctionCall(t, inputString, expectedOutputString, expectedErrorString, parseEscapeSequence)
		}
	}
}

func TestParseEscapeSequenceReturnsErrorForOnlyReverseSolidusStart(t *testing.T) {
	testIndividualParseFunctionCall(t, "\\", "\\", READ_RUNE_ERROR_EOF, parseEscapeSequence)
}

func TestParseEscapeSequenceReturnsErrorForEmptyString(t *testing.T) {
	testIndividualParseFunctionCall(t, "", "", READ_RUNE_ERROR_EOF, parseEscapeSequence)
}

func TestParseEscapeSequenceReturnsErrorForAllInvalidEscapableCharacters(t *testing.T) {
	for runeValue := rune(MIN_UNICODE_RUNE); runeValue <= MAX_UNICODE_RUNE; runeValue++ {
		switch runeValue {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't', 'u':
		default:
			inputString := string([]rune{'\\', runeValue})
			expectedOutputString := "\\"
			expectedErrorString := fmt.Sprintf("invalid escape sequence character '%c'", runeValue)
			testIndividualParseFunctionCall(t, inputString, expectedOutputString, expectedErrorString, parseEscapeSequence)
		}
	}
}

func TestParseEscapeSequenceReturnsErrorForAllInvalidHexCharactersAfterU(t *testing.T) {
	for runeValue := rune(MIN_UNICODE_RUNE); runeValue <= MAX_UNICODE_RUNE; runeValue++ {
		switch runeValue {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'A', 'B', 'C', 'D', 'E', 'F':
		default:
			inputString := string([]rune{'\\', 'u', runeValue})
			expectedOutputString := "\\u"
			expectedErrorString := fmt.Sprintf("invalid hexadecimal character in code point '%c'", runeValue)
			testIndividualParseFunctionCall(t, inputString, expectedOutputString, expectedErrorString, parseEscapeSequence)
		}
	}
}

func TestParseHexdecimalCodePointReturnsFullHexString(t *testing.T) {
	numberOfValidHexCharacters := float64(len(VALID_HEX_CHARACTERS))
	for i := float64(0); i < math.Pow(numberOfValidHexCharacters, 4); i++ {
		inputString := getFourCharacterHexStringById(int(i))
		testIndividualParseFunctionCall(t, inputString, inputString, "", parseHexadecimalCodePoint)
	}
	testIndividualParseFunctionCall(t, "00000", "0000", "", parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "0000z", "0000", "", parseHexadecimalCodePoint)
}

func getFourCharacterHexStringById(id int) string {
	n := len(VALID_HEX_CHARACTERS)
	i := id % n
	j := (id / n) % n
	k := (id / (n * n)) % n
	l := (id / (n * n * n)) % n
	return string([]rune{VALID_HEX_CHARACTERS[i], VALID_HEX_CHARACTERS[j], VALID_HEX_CHARACTERS[k], VALID_HEX_CHARACTERS[l]})
}

func TestParseHexadecimalCodePointReturnsPartialHexString(t *testing.T) {
	testIndividualParseFunctionCall(t, "wxyz", "", "invalid hexadecimal character in code point 'w'", parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "0xyz", "0", "invalid hexadecimal character in code point 'x'", parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "00yz", "00", "invalid hexadecimal character in code point 'y'", parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "000z", "000", "invalid hexadecimal character in code point 'z'", parseHexadecimalCodePoint)
}

func TestParseHexadecimalCodePointReturnsErrorForShortHexString(t *testing.T) {
	testIndividualParseFunctionCall(t, "", "", READ_RUNE_ERROR_EOF, parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "0", "0", READ_RUNE_ERROR_EOF, parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "00", "00", READ_RUNE_ERROR_EOF, parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "000", "000", READ_RUNE_ERROR_EOF, parseHexadecimalCodePoint)
}

func testIndividualParseFunctionCall(t *testing.T, inputString, expectedOutputString, expectedErrorString string, parseFunction func(ckio.PeekableRuneReader) (string, error)) {
	peekableRuneReader := createPeekableRuneReaderFromString(inputString)
	returnedString, err := parseFunction(peekableRuneReader)
	returnedErrorString := getErrorString(err)
	isExpectingError := expectedErrorString != ""
	hasReceivedError := returnedErrorString != ""
	hasMatchingErrorStrings := expectedErrorString == returnedErrorString

	if returnedString != expectedOutputString {
		t.Errorf("\"%s\" should return \"%s\", got \"%s\"", inputString, expectedOutputString, returnedString)
	}

	if isExpectingError && hasReceivedError && !hasMatchingErrorStrings {
		t.Errorf("\"%s\" should return error \"%s\", got \"%s\"", inputString, expectedErrorString, returnedErrorString)
	} else if isExpectingError && !hasReceivedError {
		t.Errorf("\"%s\" should return error \"%s\", got nil", inputString, expectedErrorString)
	} else if !isExpectingError && hasReceivedError {
		t.Errorf("\"%s\" should not return error, got \"%s\"", inputString, returnedErrorString)
	}
}

func createPeekableRuneReaderFromString(stringText string) ckio.PeekableRuneReader {
	bufioReader := bufio.NewReader(strings.NewReader(stringText))
	return ckio.NewRuneReader(bufioReader)
}

func getErrorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func TestIsHexadeciamlCharacterReturnsTrue(t *testing.T) {
	testRangeOfHexadecimalCharacters(t, '0', '9', true)
	testRangeOfHexadecimalCharacters(t, 'a', 'f', true)
	testRangeOfHexadecimalCharacters(t, 'A', 'F', true)
}

func TestIsHexadecimalCharacterReturnsFalse(t *testing.T) {
	testRangeOfHexadecimalCharacters(t, MIN_UNICODE_RUNE, '0'-1, false)
	testRangeOfHexadecimalCharacters(t, '9'+1, 'A'-1, false)
	testRangeOfHexadecimalCharacters(t, 'F'+1, 'a'-1, false)
	testRangeOfHexadecimalCharacters(t, 'f'+1, MAX_UNICODE_RUNE, false)
}

func testRangeOfHexadecimalCharacters(t *testing.T, startingCharacter, endingCharacter rune, expectedOutcome bool) {
	for runeValue := startingCharacter; runeValue != endingCharacter+1; runeValue++ {
		if isHexadecimalCharacter(runeValue) != expectedOutcome {
			t.Errorf("'%c' should return %t, got %t", runeValue, expectedOutcome, !expectedOutcome)
		}
	}
}
