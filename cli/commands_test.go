package cli

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daisy-consortium/pipeline-clientlib-go"
)

var files = []struct {
	Name, Body string
}{
	{"readme.txt", "This archive contains some text files."},
	{"fold1/gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
	{"fold1/fold2/todo.txt", "Get animal handling licence.\nWrite more examples."},
}

func createZipFile(t *testing.T) []byte {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	// Add some files to the archive.
	for _, file := range files {
		f, err := w.Create(file.Name)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
	}

	// Make sure to check the error on Close.
	err := w.Close()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	return buf.Bytes()
}

func TestDumpFiles(t *testing.T) {
	data := createZipFile(t)
	folder := filepath.Join(os.TempDir(), "pipeline_commands_test")
	err := os.MkdirAll(folder, 0755)
	visited := make(map[string]bool)
	for _, f := range files {
		visited[filepath.Clean(f.Name)] = false
	}

	defer func() {
		os.RemoveAll(folder)
	}()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	err = dumpZippedData(data, folder)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	filepath.Walk(folder, func(path string, inf os.FileInfo, err error) error {
		entry, err := filepath.Rel(folder, path)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		visited[entry] = true
		return nil
	})
	for _, f := range files {
		if !visited[filepath.Clean(f.Name)] {
			t.Errorf("%v was not visited", filepath.Clean(f.Name))
		}
	}

}

func TestQueueCommand(t *testing.T) {
	//Todo make a mechanism to mock return values from the link
	pipe := newPipelineTest(false)
	queue := []pipeline.QueueJob{
		pipeline.QueueJob{
			Id:               "job1",
			ClientPriority:   "high",
			JobPriority:      "low",
			ComputedPriority: 1.555555,
			RelativeTime:     0.577777,
			TimeStamp:        1400237879517,
		},
	}
	pipe.SetVal(queue)
	link := PipelineLink{pipeline: pipe}
	cli, err := NewCli("test", &link)
	if err != nil {
		t.Errorf("Unexpected error")
	}
	buf := make([]byte, 0)
	w := bytes.NewBuffer(buf)
	cli.Output = w
	AddQueueCommand(cli, link)
	cli.Run([]string{"queue"})
	reader := bufio.NewScanner(w)
	reader.Scan() //discard the header line
	reader.Scan()
	reader.Text()
	line := reader.Text()
	vals := strings.Split(line, "\t")
	expected := []string{
		queue[0].Id,
		fmt.Sprintf("%.2f", queue[0].ComputedPriority),
		queue[0].JobPriority,
		queue[0].ClientPriority,
		fmt.Sprintf("%.2f", queue[0].RelativeTime),
		fmt.Sprintf("%d", queue[0].TimeStamp),
	}
	for idx, _ := range vals {
		if vals[idx] != expected[idx] {
			t.Errorf("Error in displayed data %v!=%v", vals[idx], expected[idx])
		}
	}

}
