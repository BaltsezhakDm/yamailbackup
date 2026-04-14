import os
from imap_tools import MailBox, AND
from datetime import datetime, timedelta
import logging
from typing import List, Any
from utils import Config, should_process_email
from backup import upload_to_cloud

logger = logging.getLogger(__name__)

def process_mailbox(cfg: Config, last_email_date: datetime, db_conn):
    from storage import Email, save_email, email_exists

    # Go code adds -20 minutes to since date
    search_since = (last_email_date - timedelta(minutes=20)).date()

    with MailBox(cfg.imap.server, port=cfg.imap.port).login(cfg.imap.username, cfg.imap.password, cfg.imap.mailbox) as mailbox:
        # Search for messages since date
        messages = mailbox.fetch(AND(date_gte=search_since))

        unique_subjects = {}

        for msg in messages:
            # Filter by email and subject
            if should_process_email(cfg, msg.from_, msg.subject):
                # Go code keeps the one with highest SeqNum for unique subjects
                # In imap-tools, we can use msg.uid or just trust the order (usually ascending)
                unique_subjects[msg.subject] = msg

        counter = 0
        for subject, msg in unique_subjects.items():
            if email_exists(db_conn, msg.message_id):
                continue

            logger.info(f"Processing email: {msg.subject} from {msg.from_}")

            # Extract and upload attachments
            for att in msg.attachments:
                file_path = os.path.join(cfg.backup.save_path, att.filename)
                try:
                    upload_to_cloud(cfg, file_path, att.payload)
                    logger.info(f"Uploaded attachment: {att.filename}")
                except Exception as e:
                    logger.error(f"Error uploading attachment {att.filename}: {e}")

            # Save to DB
            email_to_save = Email(
                id=None,
                message_id=msg.message_id,
                subject=msg.subject,
                from_email=msg.from_,
                date=msg.date.strftime("%Y-%m-%d %H:%M:%S")
            )
            save_email(db_conn, email_to_save)
            counter += 1

        logger.info(f"Successfully processed {counter} emails")
        return counter
