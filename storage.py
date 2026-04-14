import sqlite3
from datetime import datetime
from dataclasses import dataclass
from typing import Optional

@dataclass
class Email:
    id: Optional[int]
    message_id: str
    subject: str
    from_email: str
    date: str

def init_db(db_path: str):
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    cursor.execute('''
    CREATE TABLE IF NOT EXISTS emails (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        message_id TEXT UNIQUE,
        subject TEXT,
        from_email TEXT,
        date TEXT
    );''')
    conn.commit()
    return conn

def save_email(conn, email: Email):
    cursor = conn.cursor()
    cursor.execute('''
    INSERT INTO emails (message_id, subject, from_email, date)
    VALUES (?, ?, ?, ?);
    ''', (email.message_id, email.subject, email.from_email, email.date))
    conn.commit()

def email_exists(conn, message_id: str) -> bool:
    cursor = conn.cursor()
    cursor.execute('SELECT 1 FROM emails WHERE message_id = ? LIMIT 1', (message_id,))
    return cursor.fetchone() is not None

def get_last_email_date(conn) -> datetime:
    cursor = conn.cursor()
    cursor.execute('SELECT date FROM emails ORDER BY date DESC LIMIT 1')
    row = cursor.fetchone()

    if not row or not row[0]:
        now = datetime.now()
        return datetime(now.year, now.month, now.day)

    return parse_date(row[0])

def parse_date(date_str: str) -> datetime:
    layouts = [
        "%Y-%m-%d %H:%M:%S",
        "%Y-%m-%dT%H:%M:%SZ", # Simplified ISO 8601
        "%d/%m/%Y %H:%M:%S",
        "%Y-%m-%d %H:%M:%S %z", # Handle some timezone offsets
    ]

    # Try to clean up the date string if it has too many components like in Go's sample
    # "2025-02-27 03:51:09 +0000 UTC" -> we might need to be careful

    clean_date = date_str.split(' +')[0].split(' -')[0] # Crude cleanup

    for layout in layouts:
        try:
            return datetime.strptime(date_str, layout)
        except ValueError:
            try:
                return datetime.strptime(clean_date, layout)
            except ValueError:
                continue

    # Fallback for complex Go-style dates "2025-02-27 03:51:09 +0000 UTC"
    try:
        # 2025-02-27 03:51:09 +0000 UTC
        parts = date_str.split(' ')
        if len(parts) >= 2:
            dt_str = f"{parts[0]} {parts[1]}"
            return datetime.strptime(dt_str, "%Y-%m-%d %H:%M:%S")
    except ValueError:
        pass

    raise ValueError(f"Invalid date format: {date_str}")
