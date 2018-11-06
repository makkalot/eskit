#!/usr/bin/env bash

# $1: is the service names it's a required parameter
function minikube_service_endpoints(){
   local final_str=""
	for service_name in $@; do
	    endpoint=`minikube service eskit-${service_name}-${service_name} --url | xargs | awk '{print $1;}'`
	    endpoint_grpc="${endpoint#http://}"
		endpoint_upper=`echo $service_name | awk '{print toupper($0)}'`_ENDPOINT
		export_str=`printf "export $endpoint_upper=$endpoint_grpc \n"`
		final_str="$final_str $export_str"
	done

    echo $final_str | xargs -n2
}


function wait_for_service(){
    local attempt_counter=0
    local max_attempts=10

    local service_name=$1
    local health_endpoint="http://${service_name}.local/v1/healtz"

    echo "waiting for ${service_name} to come online : ${health_endpoint}"
    until $(curl --output /dev/null --silent --fail ${health_endpoint}); do
        if [ ${attempt_counter} -eq ${max_attempts} ];then
          echo "Max attempts reached for ${service_name}"
          exit 1
        fi

        printf '.'
        attempt_counter=$(($attempt_counter+1))
        sleep 10
    done
    echo "${service_name} is online"
}