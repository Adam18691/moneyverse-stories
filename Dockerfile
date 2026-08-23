FROM python:3.11-slim
RUN apt-get update && apt-get install -y r-base ffmpeg git
RUN R -e "install.packages(c('tidyverse','data.table'), repos='https://cloud.r-project.org')"
WORKDIR /app
COPY requirements.txt.
RUN pip install -r requirements.txt
COPY..
CMD ["python", "main.py"]
