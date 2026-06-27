import concurrent.futures
import multiprocessing as mp
import os
import platform
import sys
import time


TASKS = 4
N = 3_000_000


def cpu_bound_task(n: int) -> int:
    total = 0
    for i in range(n):
        total += (i * i) % 97
    return total


def measure(label: str, runner) -> tuple[str, float, int]:
    start = time.perf_counter()
    result = runner()
    elapsed = time.perf_counter() - start
    return label, elapsed, result


def run_sequential() -> int:
    return sum(cpu_bound_task(N) for _ in range(TASKS))


def run_threading() -> int:
    with concurrent.futures.ThreadPoolExecutor(max_workers=TASKS) as executor:
        return sum(executor.map(cpu_bound_task, [N] * TASKS))


def run_multiprocessing() -> int:
    ctx = mp.get_context("spawn")
    with ctx.Pool(processes=TASKS) as pool:
        return sum(pool.map(cpu_bound_task, [N] * TASKS))


if __name__ == "__main__":
    print(f"[实测 Python {platform.python_version()}] CPU-bound 并发边界最小实验")
    print(f"platform={platform.platform()}")
    print(f"executable={sys.executable}")
    print(f"cpu_count={os.cpu_count()}, tasks={TASKS}, n_per_task={N}")
    baseline = None
    for label, runner in [
        ("sequential", run_sequential),
        ("threading-4", run_threading),
        ("multiprocessing-4", run_multiprocessing),
    ]:
        name, elapsed, result = measure(label, runner)
        if baseline is None:
            baseline = elapsed
        speedup = baseline / elapsed if elapsed > 0 else 0
        print(f"{name}: elapsed={elapsed:.3f}s, speedup_vs_sequential={speedup:.2f}x, checksum={result}")
