package util

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

var exprOperators = []rune{'+', '-', '*', '/'}

// EvaluateSimpleExpression evaluates expressions like "calc(1+calc(10/5))"
func EvaluateSimpleExpression(expr string) (int, error) {
	return evaluateSimpleExpression(expr)
}

func evaluateSimpleExpression(expr string) (int, error) {
	toEval := expr
	if strings.HasPrefix(expr, "calc(") {
		// e.g. calc(1+calc(1+1)) -> 1+calc(1+1)
		toEval = expr[5 : len(expr)-1]
	}

	// Find the operator, skipping nested calc() expressions
	var left []rune
	var right []rune
	operator := '+'
	isLeft := true
	depth := 0

	for _, r := range toEval {
		// Track depth to handle nested calc()
		if r == '(' {
			depth++
		} else if r == ')' {
			depth--
		}

		// Only split on operators at depth 0
		if depth == 0 && slices.Contains(exprOperators, r) {
			isLeft = false
			operator = r
			continue
		}

		if isLeft {
			left = append(left, r)
		} else {
			right = append(right, r)
		}
	}

	// Evaluate left operand
	var leftI int
	var err error
	if strings.HasPrefix(string(left), "calc(") {
		leftI, err = evaluateSimpleExpression(string(left))
		if err != nil {
			return -1, err
		}
	} else {
		leftI, err = strconv.Atoi(string(left))
		if err != nil {
			return -1, fmt.Errorf("invalid left operand: %s", string(left))
		}
	}

	// Evaluate right operand
	var rightI int
	if strings.HasPrefix(string(right), "calc(") {
		rightI, err = evaluateSimpleExpression(string(right))
		if err != nil {
			return -1, err
		}
	} else {
		rightI, err = strconv.Atoi(string(right))
		if err != nil {
			return -1, fmt.Errorf("invalid right operand: %s", string(right))
		}
	}

	switch operator {
	case '+':
		return leftI + rightI, nil
	case '-':
		return leftI - rightI, nil
	case '*':
		return leftI * rightI, nil
	case '/':
		if rightI == 0 {
			return -1, fmt.Errorf("cannot divide by 0")
		}
		return leftI / rightI, nil
	}

	return -1, fmt.Errorf("operator not supported")
}
