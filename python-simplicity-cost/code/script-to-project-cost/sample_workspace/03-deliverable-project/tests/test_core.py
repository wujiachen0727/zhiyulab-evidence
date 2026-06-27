import csv
from pathlib import Path
import tempfile
import unittest

from invoice_summary.core import summarize

class SummaryTest(unittest.TestCase):
    def test_summarize_amounts(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / 'invoices.csv'
            with path.open('w', newline='') as file:
                writer = csv.DictWriter(file, fieldnames=['amount'])
                writer.writeheader()
                writer.writerows([{'amount': '10'}, {'amount': '20'}])
            self.assertEqual(summarize(str(path)), 30)
