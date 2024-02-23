package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElenaGrasovskaya/gobank/types"
	"github.com/stretchr/testify/assert"
)

func TestNewAccount(t *testing.T) {
	acc, err := types.NewAccount("a", "b", "c@gmail", "111")
	assert.Nil(t, err)
	fmt.Printf("%v /n", acc)

}

func TestNewExpense(t *testing.T) {
	acc, err := types.NewExpense(1, "testa", "testb", "test", 100, time.Now())
	assert.Nil(t, err)

	fmt.Printf("%v /n", acc)
}

func TestUpdatedExpense(t *testing.T) {
	acc, err := types.UpdatedExpense(1, 1, "testa", "testb", "test", 100, time.Now())
	assert.Nil(t, err)
	fmt.Printf("%v /n", acc)
}
