#!/bin/sh
set -e
eventual migrate
# Then start the application
exec eventual start
