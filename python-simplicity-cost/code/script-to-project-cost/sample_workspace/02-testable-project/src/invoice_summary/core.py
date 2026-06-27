import csv
from pathlib import Path

def summarize(path: str) -> int:
    total = 0
    with Path(path).open() as file:
        for row in csv.DictReader(file):
            total += int(row['amount'])
    return total
