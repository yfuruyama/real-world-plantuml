FROM openjdk:8-jdk
WORKDIR /app
COPY . /app/

RUN apt update && apt install -y graphviz fonts-wqy-zenhei

EXPOSE 8080

RUN ./gradlew build

CMD ./gradlew run
