FROM nginx:alpine

WORKDIR /app

RUN apk add --update nodejs npm
RUN apk add --update apache2-utils
RUN apk add --update libwebp=1.4.0-r0
RUN apk add --update openssl>3.1.4-r1
COPY package.json package-lock.json ./

RUN npm install


COPY . .

ENTRYPOINT ["/bin/sh", "/app/scripts/deploy.sh"]
