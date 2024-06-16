FROM starky/baseimage:buildgo-latest  as build

ARG GIT_BRANCH
ARG GITHUB_SHA

ENV GOFLAGS="-mod=vendor"

ADD . /build
WORKDIR /build

RUN \
    version=$(git rev-parse --abbrev-ref HEAD)-$(git log -1 --format=%h)-$(date +%Y%m%dT%H:%M:%S) && \
    echo "version=$version" && \
    cd app && go build -o /build/gophkeeper -ldflags "-X main.revision=${version} -s -w"

FROM starky/baseimage:scratch-latest
COPY --from=build /build/gophkeeper /srv/gophkeeper
WORKDIR /srv
EXPOSE 8080

CMD ["/srv/gophkeeper"]
