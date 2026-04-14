import pytest
import sqlite3
from datetime import datetime
from storage import init_db, save_email, email_exists, get_last_email_date, parse_date, Email

@pytest.fixture
def db_conn():
    conn = init_db(":memory:")
    yield conn
    conn.close()

def test_init_db(db_conn):
    cursor = db_conn.cursor()
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='emails';")
    assert cursor.fetchone() is not None

def test_save_and_exists(db_conn):
    email = Email(
        id=None,
        message_id="test-id-123",
        subject="Test Subject",
        from_email="sender@test.com",
        date="2025-01-01 10:00:00"
    )
    save_email(db_conn, email)

    assert email_exists(db_conn, "test-id-123") is True
    assert email_exists(db_conn, "non-existent") is False

def test_get_last_email_date(db_conn):
    # Test with empty DB
    last_date = get_last_email_date(db_conn)
    now = datetime.now()
    assert last_date.year == now.year
    assert last_date.month == now.month
    assert last_date.day == now.day

    # Test with data
    emails = [
        Email(None, "id1", "s1", "f1", "2025-01-01 10:00:00"),
        Email(None, "id2", "s2", "f2", "2025-02-01 12:00:00"),
        Email(None, "id3", "s3", "f3", "2025-01-15 09:00:00"),
    ]
    for e in emails:
        save_email(db_conn, e)

    last_date = get_last_email_date(db_conn)
    assert last_date == datetime(2025, 2, 1, 12, 0, 0)

def test_parse_date():
    assert parse_date("2025-01-01 10:00:00") == datetime(2025, 1, 1, 10, 0, 0)
    assert parse_date("2025-01-01T10:00:00Z") == datetime(2025, 1, 1, 10, 0, 0)
    assert parse_date("01/01/2025 10:00:00") == datetime(2025, 1, 1, 10, 0, 0)

    # Go-style dates
    assert parse_date("2025-02-27 03:51:09 +0000 UTC") == datetime(2025, 2, 27, 3, 51, 9)
    assert parse_date("2025-02-27 03:51:09 +0300") == datetime(2025, 2, 27, 3, 51, 9)

def test_parse_date_invalid():
    with pytest.raises(ValueError):
        parse_date("invalid date")
