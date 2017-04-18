#!/bin/sh
 
# Set Heroku app id
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

source "$DIR/apiKeys.sh" # Imports MFWBOT_API_KEY variable (string)
printf "API key: %s \n" $MFWBOT_API_KEY

printf "Setting Heroku ENV vars...\n"

heroku config:set MFWBOT_API_KEY=$MFWBOT_API_KEY --app "mfw-bot"
heroku config:set GET_UPDATE_METHOD="webhook" --app "mfw-bot"

printf "\n"
