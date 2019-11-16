FROM registry.redhat.io/ubi8-minimal:latest

ENV OPERATOR=/usr/local/bin/nexus-operator \
    USER_UID=1001 \
    USER_NAME=nexus-operator

# install operator binary
COPY build/_output/bin/nexus-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
