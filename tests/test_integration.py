import pytest
import os
from imap_client import process_mailbox
from utils import Config, IMAPConfig, BackupConfig, MailConfig
from storage import init_db
from datetime import datetime, timedelta

# These tests will only run if environment variables are set
IMAP_SERVER = os.getenv("TEST_IMAP_SERVER")
IMAP_USER = os.getenv("TEST_IMAP_USER")
IMAP_PASS = os.getenv("TEST_IMAP_PASS")
CLOUD_HOST = os.getenv("TEST_CLOUD_HOST")
CLOUD_KEY = os.getenv("TEST_CLOUD_KEY")

@pytest.mark.skipif(not (IMAP_SERVER and IMAP_USER and IMAP_PASS),
                    reason="IMAP credentials not provided")
def test_real_imap_connection():
    cfg = Config(
        imap=IMAPConfig(
            server=IMAP_SERVER,
            port=993,
            username=IMAP_USER,
            password=IMAP_PASS,
            mailbox="INBOX"
        ),
        backup=BackupConfig(
            save_path="test_backups",
            host=CLOUD_HOST or "https://cloud-api.yandex.net",
            authKey=CLOUD_KEY or "dummy"
        ),
        mail=MailConfig(emails=["all"])
    )

    db_conn = init_db(":memory:")
    # Set last date to 1 day ago to fetch recent emails
    last_date = datetime.now() - timedelta(days=1)

    # We mock upload_to_cloud to avoid actually uploading to Yandex Disk unless CLOUD_KEY is real
    with pytest.MonkeyPatch().context() as m:
        if not CLOUD_KEY:
            m.setattr("imap_client.upload_to_cloud", lambda *args: None)

        count = process_mailbox(cfg, last_date, db_conn)
        print(f"Processed {count} emails from real IMAP")
        # We don't assert count > 0 because the mailbox might be empty,
        # but if it didn't crash, the connection worked.
        assert isinstance(count, int)

@pytest.mark.skipif(not (CLOUD_HOST and CLOUD_KEY),
                    reason="Cloud credentials not provided")
def test_real_cloud_upload():
    from backup import upload_to_cloud
    cfg = Config(
        backup=BackupConfig(
            host=CLOUD_HOST,
            authKey=CLOUD_KEY
        )
    )
    # Upload a small test file
    test_filename = f"test_integration_{int(datetime.now().timestamp())}.txt"
    upload_to_cloud(cfg, test_filename, b"integration test content")
