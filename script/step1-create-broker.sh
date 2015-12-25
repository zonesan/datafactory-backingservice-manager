#!/bin/bash

#BS server address
bs_server='http://localhost:41000'

#service broker info
broker_name='""'
broker_url='""'
broker_user='""'
broker_pass='""'

if [ ${broker_name} == '""' ]; then
    echo
    echo ...no broker name specified.
    echo
    exit -1
else
    echo
    echo broker name: ${broker_name}
    echo
fi

if [ ${broker_url} == '""' ]; then
    echo
    echo ...no broker url specified.
    echo
    exit -1
else
    echo
    echo broker url: ${broker_url}
    echo
fi
echo
echo "-------STEP 1---------creating service broker via ${broker_url} ..."
echo

curl -i ${bs_server}/v2/service_brokers -d "{
  \"name\": ${broker_name} ,
  \"broker_url\": ${broker_url},
  \"auth_username\": ${broker_user},
  \"auth_password\": ${broker_pass}
}" -X POST

echo
echo
