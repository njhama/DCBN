from fastapi import FastAPI, HTTPException
from typing import Dict, List
import queue

app = FastAPI()

workers = {}
task_queue = queue.Queue()
results = []

initial_tasks = [1001, 1002, 1003, 1004, 1005]
for task in initial_tasks:
    task_queue.put(task)


@app.post("/register_worker/{worker_id}")
def register_worker(worker_id: str):
    workers[worker_id] = {"status": "idle"}
    return {"message": f"Worker {worker_id} registered successfully"}


@app.get("/get_task")
def get_task():
    if task_queue.empty():
        return {"message": "No tasks available"}
    task = task_queue.get()
    return {"task": task}


@app.post("/submit_result")
def submit_result(data: Dict):
    if "worker_id" not in data or "task" not in data or "result" not in data:
        raise HTTPException(status_code=400, detail="Invalid result format")

    results.append(data)
    workers[data["worker_id"]]["status"] = "idle"
    return {"message": "Result received"}


@app.get("/results")
def get_results():
    return {"results": results}


@app.get("/workers")
def get_workers():
    return {"workers": workers}


@app.post("/add_task/{task}")
def add_task(task: int):
    task_queue.put(task)
    return {"message": f"Task {task} added"}