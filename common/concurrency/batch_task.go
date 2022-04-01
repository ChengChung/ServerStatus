package concurrency

import "sync"

type TaskResult struct {
	ID     string
	Result interface{}
}

type Task struct {
	output chan interface{}
	once   *sync.Once
	wg     *sync.WaitGroup
}

func (t *Task) Answer(reply interface{}) {
	t.once.Do(func() {
		t.output <- reply
		t.wg.Done()
	})
}

func (t *Task) AnswerWithID(id string, reply interface{}) {
	t.Answer(TaskResult{
		ID:     id,
		Result: reply,
	})
}

type BatchTask struct {
	wg   *sync.WaitGroup
	recv chan interface{}

	avail int
	size  int

	mtx *sync.Mutex
}

func NewBatchTask(size int) *BatchTask {
	wg := &sync.WaitGroup{}
	wg.Add(size)
	return &BatchTask{
		avail: size,
		size:  size,
		wg:    wg,
		recv:  make(chan interface{}, size),

		mtx: &sync.Mutex{},
	}
}

func (bt *BatchTask) WaitForFinish() []interface{} {
	bt.wg.Wait()
	close(bt.recv)

	res := make([]interface{}, 0)

	for item := range bt.recv {
		res = append(res, item)
	}

	return res
}

func (bt *BatchTask) DispatchTask() *Task {
	bt.mtx.Lock()
	defer bt.mtx.Unlock()

	if bt.avail <= 0 {
		return nil
	}

	bt.avail--

	return &Task{
		output: bt.recv,
		once:   &sync.Once{},
		wg:     bt.wg,
	}
}
