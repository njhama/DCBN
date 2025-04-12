package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	taskTopic   = "ml-tasks"
	resultTopic = "ml-results"
	brokerAddr  = "localhost:9092"
)

type Job struct {
	JobID string    `json:"job_id"`
	Data  []float64 `json:"data"`
}

type Result struct {
	WorkerID string  `json:"worker_id"`
	JobID    string  `json:"job_id"`
	Result   float64 `json:"result"`
}

var workerID string = fmt.Sprintf("worker-%d", rand.Intn(10000))

func computeJob(data []float64) float64 {
	// Example: sum of squares
	sum := 0.0
	for _, x := range data {
		sum += x * x
	}
	return sum
}

func consumeAndProcess() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddr},
		Topic:   taskTopic,
		GroupID: "go-ml-workers",
	})
	defer r.Close()

	w := &kafka.Writer{
		Addr:     kafka.TCP(brokerAddr),
		Topic:    resultTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer w.Close()

	fmt.Printf("[%s] Worker started. Listening for tasks...\n", workerID)

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println("Failed to read message:", err)
			continue
		}

		var job Job
		if err := json.Unmarshal(m.Value, &job); err != nil {
			log.Println("Failed to decode job:", err)
			continue
		}

		fmt.Printf("[%s] Received job: %+v\n", workerID, job)
		result := computeJob(job.Data)

		resultPacket := Result{
			WorkerID: workerID,
			JobID:    job.JobID,
			Result:   result,
		}

		msgBytes, _ := json.Marshal(resultPacket)
		err = w.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(job.JobID),
				Value: msgBytes,
			},
		)

		if err != nil {
			log.Println("Failed to send result:", err)
		} else {
			fmt.Printf("[%s] Sent result: %v\n", workerID, resultPacket)
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	go func() {
		for {
			fmt.Printf("[%s] Heartbeat ping\n", workerID)
			time.Sleep(10 * time.Second)
		}
	}()

	consumeAndProcess()
}
