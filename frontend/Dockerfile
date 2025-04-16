FROM node:21-alpine AS build

WORKDIR /app

COPY package.json /app/
COPY package-lock.json /app/

RUN npm install

COPY . /app/
#ARG REACT_APP_BASE_URL="http://192.168.1.25:80/api/v1"
ARG REACT_APP_BASE_URL="http://localhost:80/api/v1"
ARG IS_PROD=true
ENV IS_PROD=${IS_PROD}
ENV REACT_APP_BASE_URL=${REACT_APP_BASE_URL}

RUN npm run build


FROM nginx

COPY --from=build /app/build /usr/share/nginx/html

COPY ./nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
EXPOSE 443
