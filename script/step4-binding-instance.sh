#!/bin/bash

#BS server address
bs_server='http://localhost:41000'


#binding instance
service_instance_guid='""'


echo
echo
echo "--------STEP 4---------binding instance to application ..."
echo

if [ ${service_instance_guid} == '""' ]; then
    echo ...no service instance guid specified. please specify service_instance_guid which got from step 3.
    exit -1
else
    echo
    echo service instance guid: ${service_instance_guid}
    echo
fi

curl -i ${bs_server}/v2/service_bindings -d "{
  \"service_instance_guid\": ${service_instance_guid},
  \"app_guid\": \"72ae0608-e822-4660-a7a-ab70cf9017a7\",
  \"parameters\": {
    \"the_service_broker\": \"wants this object\"
  }
}" -X POST
