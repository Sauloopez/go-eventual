#!/bin/sh

# Ensure the nodename doesn't change, e.g. if docker restarts.
# Important because rabbitmq stores data per node name (or 'IP')
echo 'NODENAME=rabbit@localhost' > /etc/rabbitmq/rabbitmq-env.conf

if [ -z "$RABBITMQ_USER" ] || [ -z "$RABBITMQ_PASSWORD" ]; then
    echo "Error: RABBITMQ_USER and RABBITMQ_PASSWORD must be set."
    exit 1
fi

# Create Rabbitmq user
(rabbitmqctl wait --timeout 60 $RABBITMQ_PID_FILE ; \
rabbitmqctl add_user $RABBITMQ_USER $RABBITMQ_PASSWORD 2>/dev/null ; \
rabbitmqctl set_user_tags $RABBITMQ_USER administrator ; \
rabbitmqctl set_permissions -p / $RABBITMQ_USER  ".*" ".*" ".*" ; \
echo "*** User '$RABBITMQ_USER' with password '$RABBITMQ_PASSWORD' completed. ***" ; ) &

exec rabbitmq-server
