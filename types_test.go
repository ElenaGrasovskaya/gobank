package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAccount(t *testing.T) {
	acc, err := NewAccount("a", "b", "c@gmail", "111")
	assert.Nil(t, err)
	fmt.Printf("%v /n", acc)

}

func TestNewExpense(t *testing.T) {
	acc, err := NewExpense(1, "testa", "testb", "test", 100, time.Now())
	assert.Nil(t, err)

	fmt.Printf("%v /n", acc)
}

func TestUpdatedExpense(t *testing.T) {
	acc, err := UpdatedExpense(1, 1, "testa", "testb", "test", 100, time.Now())
	assert.Nil(t, err)
	fmt.Printf("%v /n", acc)
}
