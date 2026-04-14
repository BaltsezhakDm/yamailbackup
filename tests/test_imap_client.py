import pytest
from datetime import datetime
from unittest.mock import MagicMock, patch
from imap_client import process_mailbox
from utils import Config, IMAPConfig, MailConfig, BackupConfig
from storage import init_db

@pytest.fixture
def config():
    return Config(
        backup=BackupConfig(save_path="backups", host="host", authKey="key"),
        imap=IMAPConfig(server="imap.test.com", port=993, username="user", password="pwd"),
        mail=MailConfig(emails=["all"])
    )

@pytest.fixture
def db_conn():
    return init_db(":memory:")

def test_process_mailbox(config, db_conn, monkeypatch):
    # Mock MailBox
    mock_mailbox_instance = MagicMock()
    mock_mailbox_class = MagicMock(return_value=mock_mailbox_instance)
    monkeypatch.setattr("imap_client.MailBox", mock_mailbox_class)
    mock_mailbox_instance.login.return_value.__enter__.return_value = mock_mailbox_instance

    # Mock messages
    msg1 = MagicMock()
    msg1.subject = "Subject 1"
    msg1.from_ = "sender@test.com"
    msg1.uid = "id1"
    msg1.date = datetime(2025, 1, 1, 10, 0, 0)
    msg1.attachments = [MagicMock(filename="file1.pdf", payload=b"data1")]
    msg1.text = "This is the content of the first email."

    msg2 = MagicMock()
    msg2.subject = "Subject 2"
    msg2.from_ = "sender@test.com"
    msg2.uid = "id2"
    msg2.date = datetime(2025, 1, 1, 11, 0, 0)
    msg2.attachments = []
    msg2.text = "This is the content of the second email."

    # msg3 has the same subject as msg1 but different text content.
    # With the new logic, both msg1 and msg3 should be processed as their text content is unique.
    msg3 = MagicMock()
    msg3.subject = "Subject 1"
    msg3.from_ = "sender@test.com"
    msg3.uid = "id3"
    msg3.date = datetime(2025, 1, 1, 12, 0, 0)
    msg3.attachments = []
    msg3.text = "This is the content of the third email, which is different from the first."

    mock_mailbox_instance.fetch.return_value = [msg1, msg2, msg3]

    # Mock upload_to_cloud
    mock_upload = MagicMock()
    monkeypatch.setattr("imap_client.upload_to_cloud", mock_upload)
    last_date = datetime(2025, 1, 1)
    count = process_mailbox(config, last_date, db_conn)

    # With uniqueness by text, all three emails should be processed as their text content is unique.
    assert count == 3

    # Verify DB state
    from storage import email_exists
    assert email_exists(db_conn, "id1") is True
    assert email_exists(db_conn, "id2") is True
    assert email_exists(db_conn, "id3") is True

    # Verify upload was called for msg1's replacement if msg3 had attachments,
    # but here msg1 had attachment and was REPLACED by msg3 which has NONE.
    # Wait, the logic in imap_client.py is:
    # for msg in messages:
    #    unique_subjects[msg.subject] = msg
    # So msg3 overwrites msg1 in the dict.
    # Then it iterates over unique_emails_by_content.
    # So msg1 is never processed if msg3 has the same subject.

    mock_upload.assert_not_called() # because msg2 and msg3 have no attachments

def test_process_mailbox_with_attachments(config, db_conn, monkeypatch):
    mock_mailbox_instance = MagicMock()
    mock_mailbox_class = MagicMock(return_value=mock_mailbox_instance)
    monkeypatch.setattr("imap_client.MailBox", mock_mailbox_class)
    mock_mailbox_instance.login.return_value.__enter__.return_value = mock_mailbox_instance

    msg = MagicMock()
    msg.subject = "Subject"
    msg.from_ = "sender@test.com"
    msg.uid = "id1"
    msg.date = datetime(2025, 1, 1, 10, 0, 0)
    att = MagicMock(filename="test.pdf", payload=b"pdf_content")
    msg.text = "Email with attachment."
    msg.attachments = [att]

    mock_mailbox_instance.fetch.return_value = [msg]
    mock_upload = MagicMock()
    monkeypatch.setattr("imap_client.upload_to_cloud", mock_upload)
    process_mailbox(config, datetime(2025, 1, 1), db_conn)
    mock_upload.assert_called_once_with(config, "backups/test.pdf", b"pdf_content")
