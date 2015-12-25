#!/bin/bash



#BS server address
bs_server='http://localhost:41000'


#service broker info
broker_name='""'
broker_url='""'
broker_user='""'
broker_pass='""'

#create instance
#plan_guid='""'
plan_guid='""'


#binding instance
service_instance_guid='""'

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
echo "--------STEP 2----------list service plans ..."
echo

curl -i ${bs_server}/v2/service_plans

echo
echo
echo "--------STEP 3--------createing service instance ..."
echo 

if [ ${plan_guid} == '""' ]; then
    echo ...no plan guid specified.
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
echo "--------STEP 4---------binding instance to application ..."
echo

if [ ${service_instance_guid} == '""' ]; then
    echo ...no service instance guid specified.
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
