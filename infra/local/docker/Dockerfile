FROM golang:1.10

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN apt-get update && apt-get install -y ruby ruby-dev rubygems build-essential rsync
RUN gem install --no-ri --no-rdoc fpm
RUN curl -sSL https://get.docker.com/ | sh
ENTRYPOINT ["make"]
