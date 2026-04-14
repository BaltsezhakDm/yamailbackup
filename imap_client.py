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

    # Go code adds -30 minutes to since date
    search_since = (last_email_date - timedelta(minutes=30)).date()

    with MailBox(cfg.imap.server, port=cfg.imap.port).login(cfg.imap.username, cfg.imap.password, cfg.imap.mailbox) as mailbox:
        # Search for messages since date
        messages = mailbox.fetch(AND(date_gte=search_since))

        unique_emails_by_content = {}

        for msg in messages:
            # Filter by email and subject
            if should_process_email(cfg, msg.from_, msg.subject):
                email_content = msg.text if msg.text is not None else ""
                if not email_content:
                    email_content = msg.html if msg.html is not None else ""
                
                # Go code keeps the one with highest SeqNum for unique subjects
                # In imap-tools, we can use msg.uid or just trust the order (usually ascending)
                # The current implementation overwrites, effectively keeping the last one if content is identical.
                unique_emails_by_content[email_content] = msg

        counter = 0
        # Iterate over the messages that are unique by their content
        for content, msg in unique_emails_by_content.items():
            if email_exists(db_conn, msg.uid):
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
                message_id=msg.uid,
                subject=msg.subject,
                from_email=msg.from_,
                date=msg.date.strftime("%Y-%m-%d %H:%M:%S")
            )
            save_email(db_conn, email_to_save)
            counter += 1

        logger.info(f"Successfully processed {counter} emails")
        return counter
