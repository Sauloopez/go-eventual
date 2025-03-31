FROM rabbitmq:4.0-management-alpine

# Define environment variables.
ENV RABBITMQ_PID_FILE /var/lib/rabbitmq/mnesia/rabbitmq

RUN apk add --no-cache curl

# add rabbitmq_delayed_message_exchange plugin
RUN curl -L https://github.com/rabbitmq/rabbitmq-delayed-message-exchange/releases/download/v4.0.7/rabbitmq_delayed_message_exchange-v4.0.7.ez > \
    $RABBITMQ_HOME/plugins/rabbitmq_delayed_message_exchange-v4.0.7.ez

RUN chown rabbitmq:rabbitmq $RABBITMQ_HOME/plugins/rabbitmq_delayed_message_exchange-v4.0.7.ez

RUN rabbitmq-plugins enable rabbitmq_delayed_message_exchange

ADD init-rabbitmq.sh /init-rabbitmq.sh
RUN chmod +x /init-rabbitmq.sh

# Define default command
CMD ["/init-rabbitmq.sh"]
