import csv
from pathlib import Path

total = 0
with Path('invoices.csv').open() as file:
    for row in csv.DictReader(file):
        total += int(row['amount'])
print(total)
