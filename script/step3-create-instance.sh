#!/bin/bash

#BS server address
bs_server='http://localhost:41000'


#create instance
plan_guid='""'

echo
echo
echo "--------STEP 3--------createing service instance ..."
echo 

if [ ${plan_guid} == '""' ]; then
    echo ...no plan guid specified. please specify plan_guid which got from step 2.
    exit -1
else
    echo
    echo plan guid: ${plan_guid}
    echo
fi

curl -i ${bs_server}/v2/service_instances -d "{
    \"space_guid\":\"70130716-850c-41a5-a6db-d440fecacc67\",
    \"name\":\"my-service-instance\",
    \"service_plan_guid\":${plan_guid},
    \"parameters\":{
        \"the_service_broker\":\"wants this object\"
    },
    \"tags\":[
        \"accounting\",
        \"mongodb\"
    ]
}" -X POST


echo
echo
