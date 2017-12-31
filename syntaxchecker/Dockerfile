FROM gradle:4.2.1-jdk8-alpine
WORKDIR /app
COPY . /app/

USER root
RUN apk --update add graphviz
#USER gradle

EXPOSE 8080
ENV HOME /app

RUN gradle wrapper
RUN ./gradlew build
CMD ./gradlew run
