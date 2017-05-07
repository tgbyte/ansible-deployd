FROM tgbyte/ansible:2.3.0.0

RUN set -x \
  && apt-get update -qq \
  && apt-get install -y -qq --no-install-recommends \
    git \
  && apt-get clean -q \
  && rm -rf /var/lib/apt/lists/* \
  && mkdir -p /home/ansible \
  && adduser --uid 500 --disabled-login --gecos "Ansible" --no-create-home --home /home/ansible ansible

COPY out/deployd /usr/local/bin/deployd

USER ansible

ENTRYPOINT ["/usr/local/bin/deployd"]
