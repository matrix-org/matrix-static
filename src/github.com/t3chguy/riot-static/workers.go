// Copyright 2017 Michael Telatynski <7t3chguy@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	//"fmt"
	"github.com/t3chguy/riot-static/mxclient"
	"hash/fnv"
)

type JobResp interface{}
type Job interface {
	Work(w *Worker)
}

type Worker struct {
	ID     int
	client *mxclient.Client
	Queue  chan Job
	Output chan JobResp
	rooms  map[string]*mxclient.Room
}

func (w *Worker) Start() {
	for {
		job := <-w.Queue
		//fmt.Println(job)
		job.Work(w)
	}
}

type Workers struct {
	numWorkers uint32
	workers    []Worker
}

func NewWorkers(numWorkers uint32, m *mxclient.Client) *Workers {
	workers := make([]Worker, 0, numWorkers)
	for i := uint32(0); i < numWorkers; i++ {
		workers = append(workers, *NewWorker(int(i), m))
	}
	return &Workers{numWorkers, workers}
}

func mod32(a, b uint32) uint32 {
	//return uint32(math.Mod(float64(a), float64(b)))

	return a - (b * (a / b))
}

func hash(roomID string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(roomID))
	return h.Sum32()
}

func (ws *Workers) GetWorkerForRoomID(roomID string) Worker {
	workerID := mod32(hash(roomID), ws.numWorkers)

	//fmt.Println("getWorker", roomID, hash(roomID), hash(roomID), workerID, ws.numWorkers, len(ws.workers))

	return ws.workers[workerID]
}

// JobForAllWorkers sends the job to the channel of each worker.
func (ws *Workers) JobForAllWorkers(job Job) {
	for _, worker := range ws.workers {
		worker.Queue <- job
	}
}

// NewWorker instantiates a worker and their necessary channels, then starts them and returns them.
func NewWorker(id int, m *mxclient.Client) *Worker {
	worker := &Worker{
		ID:     id,
		client: m,
		Queue:  make(chan Job),
		Output: make(chan JobResp),
		rooms:  make(map[string]*mxclient.Room),
	}
	go worker.Start()
	return worker
}
