package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpr(t *testing.T) {

	tests := []struct {
		expr      string
		expected  int
		expectErr bool
	}{
		{expr: "1+1", expected: 2},
		{expr: "1-1", expected: 0},
		{expr: "2/0", expected: -1, expectErr: true},
		{expr: "10/5", expected: 2},
		{expr: "5*5", expected: 25},
		{expr: "1+calc(1+1)", expected: 3},
		{expr: "calc(1+1)+1", expected: 3},
		{expr: "1+calc(1+calc(10/5))", expected: 4},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			res, err := EvaluateSimpleExpression(fmt.Sprintf("calc(%s)", tt.expr))
			if !tt.expectErr {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expected, res)
		})
	}

}
