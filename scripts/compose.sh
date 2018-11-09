#!/usr/bin/env bash

COMPOSE_FILE=${COMPOSE_FILE:-docker-compose.yml}

# $1: is the project_name it's a required parameter
function compose_tests_golang(){
    local project_name=$1
    local compose_file=${COMPOSE_FILE}

	docker-compose -f ${compose_file} -p ${project_name} down
	docker-compose -f ${compose_file} -p ${project_name} build
	{
	    docker-compose -f ${compose_file} -p ${project_name} run --rm gotest
	} || {
		docker-compose -f ${compose_file} -p ${project_name} logs && \
		docker-compose -f ${compose_file} -p ${project_name} down && \
		exit 1
	}
	docker-compose -f ${compose_file} -p ${project_name} down

}

function compose_tests_pytest(){
    local project_name=$1

	docker-compose -p ${project_name} down
	docker-compose -p ${project_name} build
	{
	    find . -name \*.pyc -delete && \
		docker-compose -p ${project_name} run --rm pytest
	} || {
		docker-compose -p ${project_name} logs && \
		docker-compose -p ${project_name} down && \
		exit 1
	}
	docker-compose -p ${project_name} down

}


function compose_deploy(){
    local project_name=$1

	docker-compose -p ${project_name} down
	docker-compose -p ${project_name} build
	find . -name \*.pyc -delete && \
	docker-compose -p ${project_name} up

}

function compose_build(){
    local project_name=$1
    docker-compose -p ${project_name} run --rm gobuild
}