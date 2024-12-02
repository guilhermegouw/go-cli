package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/guilhermegouw/go-cli/todo"
)

var todoFileName = ".todo.json"

func main() {
	if os.Getenv("TODO_FILENAME") != "" {
		todoFileName = os.Getenv("TODO_FILENAME")
	}

	add, list, verbose, hideCompleted, complete, del := parseFlags()

	l := &todo.List{}
	if err := l.Get(todoFileName); err != nil && !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch {
	case *list:
		handleList(l, *verbose, *hideCompleted)
	case *add:
		handleAdd(l, flag.Args())
	case *complete > 0:
		handleComplete(l, *complete)
	case *del > 0:
		handleDelete(l, *del)
	default:
		fmt.Fprintln(os.Stderr, "Invalid option")
		os.Exit(1)
	}
}

func parseFlags() (*bool, *bool, *bool, *bool, *int, *int) {
	add := flag.Bool("add", false, "Add task to the To Do list.")
	list := flag.Bool("list", false, "List all tasks.")
	verbose := flag.Bool("verbose", false, "Show verbose output")
	hideCompleted := flag.Bool("hide-completed", false, "Hide completed tasks when listing")
	complete := flag.Int("complete", 0, "Item to be completed.")
	del := flag.Int("del", 0, "Item to be deleted.")

	flag.Parse()
	return add, list, verbose, hideCompleted, complete, del
}

func handleList(l *todo.List, verbose, hideCompleted bool) {
	if verbose {
		for i, task := range *l {
			fmt.Printf("  %d: %s\n  Created At: %s\n", i+1, task.Task, task.CreatedAt)
		}
	} else {
		if hideCompleted {
			filtered := &todo.List{}
			for _, task := range *l {
				if !task.Done {
					*filtered = append(*filtered, task)
				}
			}
			fmt.Print(filtered.String())
		} else {
			fmt.Print(l.String())
		}
	}
}

func handleAdd(l *todo.List, args []string) {
	tasks, err := getTask(os.Stdin, args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, task := range tasks {
		l.Add(task)
	}
	if err := l.Save(todoFileName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handleComplete(l *todo.List, id int) {
	if err := l.Complete(id); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := l.Save(todoFileName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handleDelete(l *todo.List, id int) {
	if err := l.Delete(id); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := l.Save(todoFileName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getTask(r io.Reader, args ...string) ([]string, error) {
	if len(args) > 0 {
		return []string{strings.Join(args, " ")}, nil
	}

	var tasks []string
	s := bufio.NewScanner(r)
	for s.Scan() {
		text := strings.TrimSpace(s.Text())
		if len(text) > 0 {
			tasks = append(tasks, text)
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("Task cannot be blank")
	}
	return tasks, nil
}
