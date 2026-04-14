import pytest
import yaml
import tempfile
import os
from utils import load_config, should_process_email, Config, MailConfig, BackupConfig, IMAPConfig

def test_load_config():
    config_data = {
        'backup': {
            'interval': '1h',
            'save_path': '/tmp/backup',
            'host': 'https://cloud-api.yandex.net',
            'authKey': 'secret_key'
        },
        'imap': {
            'server': 'imap.yandex.ru',
            'port': 993,
            'username': 'user@yandex.ru',
            'password': 'password',
            'mailbox': 'INBOX'
        },
        'mail': {
            'emails': ['allowed@example.com'],
            'exclude': ['blocked@example.com'],
            'subject_include': ['important'],
            'subject_exclude': ['spam']
        }
    }

    with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as tf:
        yaml.dump(config_data, tf)
        temp_name = tf.name

    try:
        cfg = load_config(temp_name)
        assert cfg.backup.interval == '1h'
        assert cfg.backup.save_path == '/tmp/backup'
        assert cfg.backup.authKey == 'secret_key'
        assert cfg.imap.server == 'imap.yandex.ru'
        assert cfg.imap.port == 993
        assert cfg.mail.emails == ['allowed@example.com']
        assert cfg.mail.subject_include == ['important']
    finally:
        os.remove(temp_name)

def test_should_process_email_by_email():
    # Case: "all" with exclusions
    cfg = Config(mail=MailConfig(emails=["all"], exclude=["blocked@test.com"]))
    assert should_process_email(cfg, "allowed@test.com", "any subject") is True
    assert should_process_email(cfg, "blocked@test.com", "any subject") is False

    # Case: Specific allowed list
    cfg = Config(mail=MailConfig(emails=["allowed1@test.com", "allowed2@test.com"]))
    assert should_process_email(cfg, "allowed1@test.com", "any") is True
    assert should_process_email(cfg, "other@test.com", "any") is False

    # Case: Case sensitivity and stripping
    cfg = Config(mail=MailConfig(emails=["Allowed@Test.Com "]))
    assert should_process_email(cfg, "  allowed@test.com", "any") is True

def test_should_process_email_by_subject():
    # Case: Subject exclusion (takes precedence)
    cfg = Config(mail=MailConfig(emails=["all"], subject_exclude=["spam", "AD"]))
    assert should_process_email(cfg, "user@test.com", "this is SPAM!") is False
    assert should_process_email(cfg, "user@test.com", "AD: buy now") is False
    assert should_process_email(cfg, "user@test.com", "Hello friend") is True

    # Case: Subject inclusion
    cfg = Config(mail=MailConfig(emails=["all"], subject_include=["Report", "Invoice"]))
    assert should_process_email(cfg, "user@test.com", "Monthly Report") is True
    assert should_process_email(cfg, "user@test.com", "Random mail") is False

    # Case: Both include and exclude
    cfg = Config(mail=MailConfig(emails=["all"], subject_include=["Report"], subject_exclude=["Draft"]))
    assert should_process_email(cfg, "user@test.com", "Final Report") is True
    assert should_process_email(cfg, "user@test.com", "Draft Report") is False

def test_empty_config_should_process():
    # Default is no emails, so nothing should be processed unless it's "all"
    cfg = Config(mail=MailConfig(emails=[]))
    assert should_process_email(cfg, "user@test.com", "subject") is False

    cfg = Config(mail=MailConfig(emails=["all"]))
    assert should_process_email(cfg, "user@test.com", "subject") is True
