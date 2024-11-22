package main

import (
	"bytes"
	"testing"
)

func TestCountWords(t *testing.T) {
	b := bytes.NewBufferString("word1 word2 word3 word4\n")

	exp := 4
	res, err := count(b, false, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if exp != res {
		t.Errorf("Expected %d got %d instead", exp, res)
	}
}

func TestCountLines(t *testing.T) {
	b := bytes.NewBufferString("word1 \nline2 word2 \nline3 word3 word4")

	exp := 3
	res, err := count(b, false, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if exp != res {
		t.Errorf("Expected %d got %d instead", exp, res)
	}
}

func TestCountBytes(t *testing.T) {
	b := bytes.NewBufferString("hello\n")

	exp := 6
	res, err := count(b, false, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if exp != res {
		t.Errorf("Expected %d got %d instead", exp, res)
	}
}

func TestCountWithNilReader(t *testing.T) {
	_, err := count(nil, false, false)
	if err == nil {
		t.Error("Expected error with nil reader, got nil")
	}
}
