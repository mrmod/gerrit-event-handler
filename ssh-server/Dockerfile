FROM ubuntu:latest

WORKDIR /app
RUN apt-get update && apt-get install -y git openssh-server

RUN mkdir /git-root
RUN groupadd git --gid 5000
RUN useradd git -m -d /git -s /bin/git-shell -c "Git Remote User" --uid 5000 --gid 5000

RUN mkdir /git/.ssh
RUN chown -R git:git /git-root
RUN chown -R git:git /git



COPY ./start.sh /app/start.sh
RUN chmod +x /app/start.sh

RUN echo "PasswordAuthentication no" >> /etc/ssh/sshd_config.d/no_password_authentication
RUN echo "LogLevel DEBUG" >> /etc/ssh/sshd_config.d/log_level_debug

ENTRYPOINT ["/app/start.sh"]