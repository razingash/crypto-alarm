FROM python:3.9-slim

RUN apt-get update && apt-get install -y libpq-dev

WORKDIR /app

COPY requirements/dev.txt requirements.txt
COPY . /app/

RUN pip install -r requirements.txt
