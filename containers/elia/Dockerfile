FROM ubuntu:latest

RUN apt update -y
RUN apt install -y --no-install-recommends pipx git-core
RUN pipx ensurepath
RUN pipx install git+https://github.com/darrenburns/elia
ENV OPENAI_API_KEY=none
ENV PATH=$PATH:/root/.local/bin

CMD ["elia"]
