package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/guilhermegouw/go-cli/todo"
)

const (
	binName = "todo"
)

func getCreatedAtFromFile(t *testing.T, taskName string) string {
	t.Helper()

	fileName := os.Getenv("TODO_FILENAME")
	if fileName == "" {
		t.Fatal("TODO_FILENAME environment variable not set")
	}

	file, err := os.Open(fileName)
	if err != nil {
		t.Fatalf("failed to open task file: %v", err)
	}
	defer file.Close()

	var tasks []todo.Task
	if err := json.NewDecoder(file).Decode(&tasks); err != nil {
		t.Fatalf("failed to decode task file: %v", err)
	}

	for _, task := range tasks {
		if task.Task == taskName {
			return task.CreatedAt.Format(time.RFC3339Nano)
		}
	}
	t.Fatalf("task %q not found in task file", taskName)
	return ""
}

func TestMain(m *testing.M) {
	tmpFileName := fmt.Sprintf(".todo.%d.json", time.Now().UnixNano())
	os.Setenv("TODO_FILENAME", tmpFileName)

	fmt.Println("Building tool...")
	build := exec.Command("go", "build", "-o", binName)

	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot build tool %s: %s", binName, err)
		os.Exit(1)
	}

	fmt.Println("Running tests...")
	result := m.Run()

	fmt.Println("Cleaning up...")
	os.Remove(binName)
	os.Remove(tmpFileName)

	os.Exit(result)
}

func TestTodoCLI(t *testing.T) {
	task1 := "Test task number 1"
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cmdPath := filepath.Join(dir, binName)

	t.Run("AddNewTaskFromArguments", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add", task1)

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("ListWithVerboseFlag", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list", "--verbose")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		createdAtRaw := getCreatedAtFromFile(t, task1)
		createdAt, err := time.Parse(time.RFC3339Nano, createdAtRaw)
		if err != nil {
			t.Fatalf("Failed to parse CreatedAt: %v", err)
		}

		expect := fmt.Sprintf("  1: %s\n  Created At: %s\n", task1, createdAt)

		if expect != string(out) {
			t.Errorf("Verbose output mismatch.\nExpected:\n%q\nGot:\n%q", expect, string(out))
		}
	})

	task2 := "test task number 2"
	t.Run("AddNewTaskFromSTDIN", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add")
		cmdStdIn, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		io.WriteString(cmdStdIn, task2)
		cmdStdIn.Close()

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ListTasks", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		expect := fmt.Sprintf("  1: %s\n  2: %s\n", task1, task2)

		if expect != string(out) {
			t.Errorf("Expected %q Got %q instead\n", expect, string(out))
		}
	})

	t.Run("CompleteTask", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-complete", "1")
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}

		cmd = exec.Command(cmdPath, "-list")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		expect := fmt.Sprintf("X 1: %s\n  2: %s\n", task1, task2)
		if expect != string(out) {
			t.Errorf("Expected %q Got %q instead\n", expect, string(out))
		}
	})

	t.Run("ListHiddingCompletedTasks", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list", "--hide-completed")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		expect := fmt.Sprintf("  1: %s\n", task2)
		if expect != string(out) {
			t.Errorf("Expected %q, got %q instead\n", expect, string(out))
		}
	})

	t.Run("DeleteTask", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-del", "1")
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}

		cmd = exec.Command(cmdPath, "-list")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		expect := fmt.Sprintf("  1: %s\n", task2)
		if expect != string(out) {
			t.Errorf("Expected %q Got %q instead\n", expect, string(out))
		}
		cmd = exec.Command(cmdPath, "-del", "1")
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}

		cmd = exec.Command(cmdPath, "-list")
		out, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}
		expect = fmt.Sprintf("")
		if expect != string(out) {
			t.Errorf("Expected %q Got %q instead\n", expect, string(out))
		}
	})
	t.Run("AddultipleTasksFromSTDIN", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add")
		cmdStdIn, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}

		tasks := []string{"multiline task 1", "multiline task 2", "multiline task 3"}
		for _, task := range tasks {
			io.WriteString(cmdStdIn, task+"\n")
		}
		cmdStdIn.Close()

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}

		cmd = exec.Command(cmdPath, "-list")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		expect := fmt.Sprintf("  1: %s\n  2: %s\n  3: %s\n",
			tasks[0], tasks[1], tasks[2])
		if expect != string(out) {
			t.Errorf("Expected %q Got %q instead\n", expect, string(out))
		}
	})
}
