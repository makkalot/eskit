FROM python:3

WORKDIR /usr/src/app

COPY ./pyservices/store/requirements.txt ./pyservices/store/requirements.txt
RUN pip install -r ./pyservices/store/requirements.txt

COPY ./Makefile ./Makefile
COPY ./pyservices/store ./pyservices/store
COPY ./pyservices/generated ./pyservices/generated/

ENV PYTHONPATH=/usr/src/app
ENV PYTHONPATH=$PYTHONPATH:/usr/src/app/pyservices
ENV PYTHONPATH=$PYTHONPATH:/usr/src/app/pyservices/generated

CMD [ "python", "./pyservices/store/main.py" ]