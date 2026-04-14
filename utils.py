import yaml
from dataclasses import dataclass, field
from typing import List

@dataclass
class BackupConfig:
    interval: str = ""
    save_path: str = ""
    clientId: str = ""
    clientSecret: str = ""
    host: str = ""
    authKey: str = ""

@dataclass
class IMAPConfig:
    server: str = ""
    port: int = 0
    username: str = ""
    password: str = ""
    mailbox: str = "INBOX"

@dataclass
class MailConfig:
    emails: List[str] = field(default_factory=list)
    exclude: List[str] = field(default_factory=list)
    subject_include: List[str] = field(default_factory=list)
    subject_exclude: List[str] = field(default_factory=list)

@dataclass
class Config:
    backup: BackupConfig = field(default_factory=BackupConfig)
    imap: IMAPConfig = field(default_factory=IMAPConfig)
    mail: MailConfig = field(default_factory=MailConfig)

def load_config(filename: str) -> Config:
    with open(filename, 'r') as f:
        data = yaml.safe_load(f) or {}

    backup_data = data.get('backup', {})
    imap_data = data.get('imap', {})
    mail_data = data.get('mail', {})

    return Config(
        backup=BackupConfig(**backup_data),
        imap=IMAPConfig(**imap_data),
        mail=MailConfig(**mail_data)
    )

def should_process_email(config: Config, from_email: str, subject: str) -> bool:
    if not _should_process_by_email(config.mail, from_email):
        return False
    return _should_process_by_subject(config.mail, subject)

def _should_process_by_email(mail_cfg: MailConfig, email: str) -> bool:
    e = email.strip().lower()

    if len(mail_cfg.emails) == 1 and mail_cfg.emails[0].strip().lower() == "all":
        for excl in mail_cfg.exclude:
            if excl.strip().lower() == e:
                return False
        return True

    for allowed in mail_cfg.emails:
        if allowed.strip().lower() == e:
            return True

    return False

def _should_process_by_subject(mail_cfg: MailConfig, subject: str) -> bool:
    s = subject.lower()

    for excl in mail_cfg.subject_exclude:
        excl = excl.strip().lower()
        if excl and excl in s:
            return False

    if not mail_cfg.subject_include:
        return True

    for incl in mail_cfg.subject_include:
        incl = incl.strip().lower()
        if incl and incl in s:
            return True

    return False
