#!/bin/bash

#BS server address
bs_server='http://localhost:41000'

echo
echo "--------STEP 2----------list service plans ..."
echo

curl -i ${bs_server}/v2/service_plans

echo
echo
