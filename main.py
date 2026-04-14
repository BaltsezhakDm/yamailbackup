import argparse
import logging
import os
import sys
from utils import load_config
from storage import init_db, get_last_email_date
from imap_client import process_mailbox

# Setup logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

def run():
    parser = argparse.ArgumentParser(description='Yandex Email Backup')
    parser.add_argument('--config', default='config/config.yaml', help='Path to config file')
    args = parser.parse_args()

    config_path = args.config
    if not os.path.exists(config_path):
        config_path = '/app/config/config.yaml'
        if not os.path.exists(config_path):
            logger.error(f"Config file not found at {args.config} or {config_path}")
            sys.exit(1)

    try:
        config = load_config(config_path)
    except Exception as e:
        logger.error(f"Error loading config: {e}")
        sys.exit(1)

    logger.info("Starting yamailbackup (Python version)...")
    logger.info(f"Backup interval: {config.backup.interval}")
    logger.info(f"Save path: {config.backup.save_path}")

    db_path = "messages.db"
    try:
        conn = init_db(db_path)
    except Exception as e:
        logger.error(f"Error initializing database: {e}")
        sys.exit(1)

    try:
        last_email_date = get_last_email_date(conn)
        logger.info(f"Last email date from DB: {last_email_date}")

        process_mailbox(config, last_email_date, conn)
    except Exception as e:
        logger.error(f"Error during backup process: {e}")
        conn.close()
        sys.exit(1)
    finally:
        conn.close()

if __name__ == "__main__":
    run()
