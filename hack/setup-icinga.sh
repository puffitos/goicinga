#!/usr/bin/bash
docker network create icinga

# CA
docker run --rm \
	-h icinga-master \
	-v icinga-master:/data \
	-e ICINGA_MASTER=1 \
	icinga/icinga2 \
	cat /var/lib/icinga2/certs/ca.crt > icinga-ca.crt

# Ticket
docker run --rm \
	-h icinga-master \
	-v icinga-master:/data \
	-e ICINGA_MASTER=1 \
	icinga/icinga2 \
	icinga2 daemon -C
docker run --rm \
	-h icinga-master \
	-v icinga-master:/data \
	-e ICINGA_MASTER=1 \
	icinga/icinga2 \
	icinga2 pki ticket --cn icinga-agent > icinga-agent.ticket

# Master
docker run --rm -d \
	--network icinga \
	--name icinga-master \
	-h icinga-master \
	-p 5665:5665 \
	-v icinga-master:/data \
	-e ICINGA_MASTER=1 \
	icinga/icinga2

# Agent
docker run --rm -d \
	--network icinga \
	-h icinga-agent \
	-v icinga-agent:/data \
	-e ICINGA_ZONE=icinga-agent \
	-e ICINGA_ENDPOINT=icinga-master,icinga-master,5665 \
	-e ICINGA_CACERT="$(< icinga-ca.crt)" \
	-e ICINGA_TICKET="$(< icinga-agent.ticket)" \
	icinga/icinga2
