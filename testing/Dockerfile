FROM php:8.3-cli-alpine
ARG PROTOC_VERSION=25.1
ARG PROTOC_ZIP="protoc-$PROTOC_VERSION-linux-x86_64.zip"

RUN apk add curl unzip bash

# install protoc
RUN curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/$PROTOC_ZIP
RUN unzip -o $PROTOC_ZIP -d /usr/local bin/protoc
RUN unzip -o $PROTOC_ZIP -d /usr/local 'include/*'
RUN rm -f $PROTOC_ZIP

# Install Composer
RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer
RUN chmod +x /usr/local/bin/composer

#ADD volume /tests
#ADD protoc-gen-php8 /tests/bin/protoc-gen-php8

WORKDIR /tests

CMD ["/bin/bash", "/tests/entrypoint.sh"]