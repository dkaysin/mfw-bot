#!/bin/sh
 
# Set Heroku app id
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source "$DIR/apiKeys.sh"

# source config/apiKeys.sh

printf "Setting Heroku ENV vars...\n"

heroku config:set MFWBOT_API_KEY=$MFWBOT_API_KEY --app $APP

printf "\n"
