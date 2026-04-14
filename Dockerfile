FROM python:3.12-alpine

WORKDIR /app

# Install system dependencies
RUN apk add --no-cache tzdata curl bash busybox-extras

# Copy requirements and install python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application
COPY . .

# Set permissions
RUN chmod +x /app/main.py

# Copy crontab and install it
COPY crontab /etc/crontabs/root

# Start crond in foreground with logging
CMD ["crond", "-f", "-l", "8"]
