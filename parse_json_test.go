package ckjson

import (
	"bufio"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/CKachur/ckio"
)

const minControlRune rune = 0
const maxControlRune rune = 0x1F
const minUnicodeRune rune = 0
const maxUnicodeRune rune = 0x10FFFF
const readRuneErrorEof string = "could not read rune: EOF"

var validHexCharacters []rune = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'A', 'B', 'C', 'D', 'E', 'F'}

func TestParseTrueReturnsValidTrue(t *testing.T) {
	testIndividualParseFunctionCall(t, "true", "true", "", parseTrue)
}

func TestParseFalseReturnsValidFalse(t *testing.T) {
	testIndividualParseFunctionCall(t, "false", "false", "", parseFalse)
}

func TestParseNullReturnsValidNull(t *testing.T) {
	testIndividualParseFunctionCall(t, "null", "null", "", parseNull)
}

func TestParseNumberValid(t *testing.T) {
	validNumbers := []string{
		"0", "0e123", "0E123", "0e+456", "0e-7890", "0E+012", "0E-3456789",
		"0.012e123", "0.345E123", "0.678e+456", "0.901e-7890", "0.234E+012", "0.567E-3456789",
		"1", "2e123", "3E123", "4e+456", "5e-7890", "6E+012", "7E-3456789",
		"8.012e123", "9.345E123", "1.678e+456", "2.901e-7890", "3.234E+012", "4.567E-3456789",
		"112", "234e123", "356E123", "478e+456", "590e-7890", "601E+012", "723E-3456789",
		"845.012e123", "967.345E123", "189.678e+456", "201.901e-7890", "323.234E+012", "445.567E-3456789"}
	for _, validNumberString := range validNumbers {
		negativeValidNumberString := fmt.Sprintf("-%s", validNumberString)
		testIndividualParseFunctionCall(t, validNumberString, validNumberString, "", parseNumber)
		testIndividualParseFunctionCall(t, negativeValidNumberString, negativeValidNumberString, "", parseNumber)
	}
}

func TestParseExponentReturnsValidExponentNoSign(t *testing.T) {
	testIndividualParseFunctionCall(t, "e0123456789", "e0123456789", "", parseExponent)
	testIndividualParseFunctionCall(t, "E1", "E1", "", parseExponent)
}

func TestParseExponentReturnsValidExponentWithSign(t *testing.T) {
	testIndividualParseFunctionCall(t, "e+0123456789", "e+0123456789", "", parseExponent)
	testIndividualParseFunctionCall(t, "e-0123456789", "e-0123456789", "", parseExponent)
	testIndividualParseFunctionCall(t, "E+1", "E+1", "", parseExponent)
	testIndividualParseFunctionCall(t, "E-1", "E-1", "", parseExponent)
}

func TestParseExponentReturnsInvalidExponentNoE(t *testing.T) {
	testIndividualParseFunctionCall(t, "1", "", "expected 'e' or 'E', found '1'", parseExponent)
}

func TestParseExponentValid(t *testing.T) {
	digits := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for _, i := range digits {
		testIndividualParseExponentValid(t, fmt.Sprintf("e%d", i))
		testIndividualParseExponentValid(t, fmt.Sprintf("e+%d", i))
		testIndividualParseExponentValid(t, fmt.Sprintf("e-%d", i))
		testIndividualParseExponentValid(t, fmt.Sprintf("E%d", i))
		testIndividualParseExponentValid(t, fmt.Sprintf("E+%d", i))
		testIndividualParseExponentValid(t, fmt.Sprintf("E-%d", i))
		for _, j := range digits {
			testIndividualParseExponentValid(t, fmt.Sprintf("e%d%d", i, j))
			testIndividualParseExponentValid(t, fmt.Sprintf("e+%d%d", i, j))
			testIndividualParseExponentValid(t, fmt.Sprintf("e-%d%d", i, j))
			testIndividualParseExponentValid(t, fmt.Sprintf("E%d%d", i, j))
			testIndividualParseExponentValid(t, fmt.Sprintf("E+%d%d", i, j))
			testIndividualParseExponentValid(t, fmt.Sprintf("E-%d%d", i, j))
			for _, k := range digits {
				testIndividualParseExponentValid(t, fmt.Sprintf("e%d%d%d", i, j, k))
				testIndividualParseExponentValid(t, fmt.Sprintf("e+%d%d%d", i, j, k))
				testIndividualParseExponentValid(t, fmt.Sprintf("e-%d%d%d", i, j, k))
				testIndividualParseExponentValid(t, fmt.Sprintf("E%d%d%d", i, j, k))
				testIndividualParseExponentValid(t, fmt.Sprintf("E+%d%d%d", i, j, k))
				testIndividualParseExponentValid(t, fmt.Sprintf("E-%d%d%d", i, j, k))
			}
		}
	}
}

func TestParseExponentInvalidNoE(t *testing.T) {
	for i := minUnicodeRune; i <= maxUnicodeRune; i++ {
		if i != 'e' && i != 'E' {
			testIndividualParseFunctionCall(t, string(i), "", fmt.Sprintf("expected 'e' or 'E', found '%c'", i), parseExponent)
		}
	}
}

func TestParseExponentInvalidWithE(t *testing.T) {
	for i := minUnicodeRune; i <= maxUnicodeRune; i++ {
		if i != '+' && i != '-' && !isDigit(i) {
			lowercaseEString := fmt.Sprintf("e%c", i)
			uppercaseEString := fmt.Sprintf("E%c", i)
			expectedErrorString := "expected '+', '-', or digit in exponent"
			testIndividualParseFunctionCall(t, lowercaseEString, "e", expectedErrorString, parseExponent)
			testIndividualParseFunctionCall(t, uppercaseEString, "E", expectedErrorString, parseExponent)
		}
	}
}

func TestParseExponentInvalidWithEAndSign(t *testing.T) {
	for i := minUnicodeRune; i <= maxUnicodeRune; i++ {
		if !isDigit(i) {
			lowercaseEPlusString := fmt.Sprintf("e+%c", i)
			lowercaseEMinusString := fmt.Sprintf("e-%c", i)
			uppercaseEPlusString := fmt.Sprintf("E+%c", i)
			uppercaseEMinusString := fmt.Sprintf("E-%c", i)
			expectedErrorString := "expected digit in exponent"
			testIndividualParseFunctionCall(t, lowercaseEPlusString, "e+", expectedErrorString, parseExponent)
			testIndividualParseFunctionCall(t, lowercaseEMinusString, "e-", expectedErrorString, parseExponent)
			testIndividualParseFunctionCall(t, uppercaseEPlusString, "E+", expectedErrorString, parseExponent)
			testIndividualParseFunctionCall(t, uppercaseEMinusString, "E-", expectedErrorString, parseExponent)
		}
	}
}

func testIndividualParseExponentValid(t *testing.T, exponent string) {
	testIndividualParseFunctionCall(t, exponent, exponent, "", parseExponent)
}

func TestParsePreDecimalDigitsValidZero(t *testing.T) {
	testIndividualParseFunctionCall(t, "0", "0", "", parsePreDecimalDigits)
	for i := minUnicodeRune; i <= maxUnicodeRune; i++ {
		zeroString := fmt.Sprintf("0%c", i)
		testIndividualParseFunctionCall(t, zeroString, "0", "", parsePreDecimalDigits)
	}
}

func TestParsePreDecimalDigitsValidNonZeroDigits(t *testing.T) {
	digits := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for _, i := range digits[1:] {
		singleDigitString := fmt.Sprintf("%d", i)
		testIndividualParseFunctionCall(t, singleDigitString, singleDigitString, "", parsePreDecimalDigits)
		for _, j := range digits {
			doubleDigitString := fmt.Sprintf("%s%d", singleDigitString, j)
			testIndividualParseFunctionCall(t, doubleDigitString, doubleDigitString, "", parsePreDecimalDigits)
			for _, k := range digits {
				tripleDigitString := fmt.Sprintf("%s%d", doubleDigitString, k)
				testIndividualParseFunctionCall(t, tripleDigitString, tripleDigitString, "", parsePreDecimalDigits)
			}
		}
	}
}

func TestParsePreDecimalDigitsInvalid(t *testing.T) {
	for i := minUnicodeRune; i <= maxUnicodeRune; i++ {
		if !isDigit(i) {
			testIndividualParseFunctionCall(t, string(i), "", fmt.Sprintf("expected digit, found '%c'", i), parsePreDecimalDigits)
		}
	}
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func TestParseStringReturnsErrorForNonQuoteStart(t *testing.T) {
	for i := minUnicodeRune; i <= maxUnicodeRune; i++ {
		if i != '"' {
			inputString := string(i)
			expectedErrorString := fmt.Sprintf("expected '\"' to begin string, found '%c'", i)
			testIndividualParseFunctionCall(t, inputString, "", expectedErrorString, parseString)
		}
	}
}

func TestParseStringReturnsValidString(t *testing.T) {
	testIndividualParseFunctionCall(t, "\"hello, json\"", "\"hello, json\"", "", parseString)
	testIndividualParseFunctionCall(t, "\"\"", "\"\"", "", parseString)
	testIndividualParseFunctionCall(t, "\"fire cat ㊋🔥😼🔥㊋\"", "\"fire cat ㊋🔥😼🔥㊋\"", "", parseString)
}

func TestParseStringReturnsValidEscapeSequenceString(t *testing.T) {
	testIndividualParseFunctionCall(t, "\"\\\"\\\\\\/\\b\\f\\n\\r\\t\"", "\"\\\"\\\\\\/\\b\\f\\n\\r\\t\"", "", parseString)
	testIndividualParseFunctionCall(t, "\"\\u0000\\u0000\"", "\"\\u0000\\u0000\"", "", parseString)
	numberOfValidHexCharacters := float64(len(validHexCharacters))
	for i := float64(0); i < math.Pow(numberOfValidHexCharacters, 4); i++ {
		inputString := fmt.Sprintf("\"\\u%s\"", getFourCharacterHexStringById(int(i)))
		testIndividualParseFunctionCall(t, inputString, inputString, "", parseString)
	}
}

func TestParseStringReturnsErrorForControlCharacters(t *testing.T) {
	for i := minControlRune; i <= maxControlRune; i++ {
		inputString := string([]rune{'"', i})
		expectedOutputString := "\""
		expectedErrorString := fmt.Sprintf("unexpected control character 0x%X", i)
		testIndividualParseFunctionCall(t, inputString, expectedOutputString, expectedErrorString, parseString)
	}
}

func TestIsControlCharacterReturnsTrue(t *testing.T) {
	for i := minControlRune; i <= maxControlRune; i++ {
		if !isControlCharacter(i) {
			t.Errorf("control character 0x%X should return true, got false", i)
		}
	}
}

func TestIsControlCharacterReturnsFalse(t *testing.T) {
	for i := maxControlRune + 1; i <= maxUnicodeRune; i++ {
		if isControlCharacter(i) {
			t.Errorf("control character 0x%X should return false, got true", i)
		}
	}
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

	numberOfValidHexCharacters := float64(len(validHexCharacters))
	for i := float64(0); i < math.Pow(numberOfValidHexCharacters, 4); i++ {
		inputString := fmt.Sprintf("\\u%s", getFourCharacterHexStringById(int(i)))
		testIndividualParseFunctionCall(t, inputString, inputString, "", parseEscapeSequence)
	}
}

func TestParseEscapeSequenceReturnsErrorForNonReverseSolidusStart(t *testing.T) {
	for runeValue := rune(minUnicodeRune); runeValue <= maxUnicodeRune; runeValue++ {
		if runeValue != '\\' {
			inputString := string(runeValue)
			expectedOutputString := ""
			expectedErrorString := fmt.Sprintf("expected '\\' at beginning of escape sequence, found '%c'", runeValue)
			testIndividualParseFunctionCall(t, inputString, expectedOutputString, expectedErrorString, parseEscapeSequence)
		}
	}
}

func TestParseEscapeSequenceReturnsErrorForOnlyReverseSolidusStart(t *testing.T) {
	testIndividualParseFunctionCall(t, "\\", "\\", readRuneErrorEof, parseEscapeSequence)
}

func TestParseEscapeSequenceReturnsErrorForEmptyString(t *testing.T) {
	testIndividualParseFunctionCall(t, "", "", readRuneErrorEof, parseEscapeSequence)
}

func TestParseEscapeSequenceReturnsErrorForAllInvalidEscapableCharacters(t *testing.T) {
	for runeValue := rune(minUnicodeRune); runeValue <= maxUnicodeRune; runeValue++ {
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
	for runeValue := rune(minUnicodeRune); runeValue <= maxUnicodeRune; runeValue++ {
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
	numberOfValidHexCharacters := float64(len(validHexCharacters))
	for i := float64(0); i < math.Pow(numberOfValidHexCharacters, 4); i++ {
		inputString := getFourCharacterHexStringById(int(i))
		testIndividualParseFunctionCall(t, inputString, inputString, "", parseHexadecimalCodePoint)
	}
	testIndividualParseFunctionCall(t, "00000", "0000", "", parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "0000z", "0000", "", parseHexadecimalCodePoint)
}

func getFourCharacterHexStringById(id int) string {
	n := len(validHexCharacters)
	i := id % n
	j := (id / n) % n
	k := (id / (n * n)) % n
	l := (id / (n * n * n)) % n
	return string([]rune{validHexCharacters[i], validHexCharacters[j], validHexCharacters[k], validHexCharacters[l]})
}

func TestParseHexadecimalCodePointReturnsPartialHexString(t *testing.T) {
	testIndividualParseFunctionCall(t, "wxyz", "", "invalid hexadecimal character in code point 'w'", parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "0xyz", "0", "invalid hexadecimal character in code point 'x'", parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "00yz", "00", "invalid hexadecimal character in code point 'y'", parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "000z", "000", "invalid hexadecimal character in code point 'z'", parseHexadecimalCodePoint)
}

func TestParseHexadecimalCodePointReturnsErrorForShortHexString(t *testing.T) {
	testIndividualParseFunctionCall(t, "", "", readRuneErrorEof, parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "0", "0", readRuneErrorEof, parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "00", "00", readRuneErrorEof, parseHexadecimalCodePoint)
	testIndividualParseFunctionCall(t, "000", "000", readRuneErrorEof, parseHexadecimalCodePoint)
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
	testRangeOfHexadecimalCharacters(t, minUnicodeRune, '0'-1, false)
	testRangeOfHexadecimalCharacters(t, '9'+1, 'A'-1, false)
	testRangeOfHexadecimalCharacters(t, 'F'+1, 'a'-1, false)
	testRangeOfHexadecimalCharacters(t, 'f'+1, maxUnicodeRune, false)
}

func testRangeOfHexadecimalCharacters(t *testing.T, startingCharacter, endingCharacter rune, expectedOutcome bool) {
	for runeValue := startingCharacter; runeValue != endingCharacter+1; runeValue++ {
		if isHexadecimalCharacter(runeValue) != expectedOutcome {
			t.Errorf("'%c' should return %t, got %t", runeValue, expectedOutcome, !expectedOutcome)
		}
	}
}
