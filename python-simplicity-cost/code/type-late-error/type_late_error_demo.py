from typing import TypedDict


class UserRecord(TypedDict):
    name: str
    age: int


def normalize_name_untyped(record):
    return record["name"].strip().lower()


def normalize_name_typed(record: UserRecord) -> str:
    return record["name"].strip().lower()


untyped_records = [
    {"name": "Ada", "age": 36},
    {"name": "Grace", "age": 85},
    {"name": None, "age": 41},
]

typed_records: list[UserRecord] = [
    {"name": "Ada", "age": 36},
    {"name": "Grace", "age": 85},
    {"name": None, "age": 41},
]


if __name__ == "__main__":
    print("[实测 Python 3.9.6] 无类型检查路径：前两个输入正常，第三个输入到运行时才报错")
    for index, record in enumerate(untyped_records, start=1):
        try:
            print(f"case {index}: {normalize_name_untyped(record)}")
        except Exception as err:
            print(f"case {index}: {type(err).__name__}: {err}")
