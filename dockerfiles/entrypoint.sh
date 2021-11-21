#!/bin/bash
set -e

envsubst '
  $YKTR2_ADDR
  $YKTR2_PORT
  $YKTR2_TAEM
  $YKTR2_PER_PAGE
  $YKTR2_SESSION_SECRET
  $YKTR2_OAUTH2_CLIENT_ID
  $YKTR2_OAUTH2_CLIENT_SECRET
  $YKTR2_OAUTH2_REDIRECT_HOST
' < /yktr2.toml.template > /yktr2.toml

exec /yktr2 -config /yktr2.toml
