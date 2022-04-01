package concurrency

import "testing"

func run_task(task *Task, id int) {
	task.Answer(id)
}

func TestBatchTask(t *testing.T) {
	size := 100000

	batch := NewBatchTask(size)

	idx := 0
	for {
		task := batch.DispatchTask()
		if task == nil {
			break
		}
		go run_task(task, idx)
		idx++
	}

	result := batch.WaitForFinish()

	if len(result) == size {
		t.Log("succ")
	} else {
		t.Fatal("fail")
	}
}
