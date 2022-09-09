#!/usr/bin/env bash

########################
# include the magic
########################
. ./demo-magic.sh

########################
# Configure the options
########################

#
# speed at which to simulate typing. bigger num = faster
#
# TYPE_SPEED=20

#
# custom prompt
#
# see http://www.tldp.org/HOWTO/Bash-Prompt-HOWTO/bash-prompt-escape-sequences.html for escape sequences
#
DEMO_PROMPT="${CYAN}\W ${GREEN}➜ "

CNBI_MANIFEST="${MANIFEST:-op1st-ds.yaml}"
CLEANUP="${CLEANUP:-True}"

# text color
# DEMO_CMD_COLOR=$BLACK

# hide the evidence
clear

# Create a CustomNBImage of buildType 'GitRepository'
pe "cat ${CNBI_MANIFEST}"
pe "oc apply -f ${CNBI_MANIFEST}"
pe "oc get customnbimage"

# Show how a PipelineRun has been created
pe "tkn pr ls"

# Follow the logs of the PipelineRun
# ctl + c support: ctl + c to stop long-running process and continue demo
pe "tkn pr logs -f -L"

# Show the final result
pe "oc get customnbimage"
pei "tkn pr ls"
pe "oc get is"

# Cleanup
if [ "$CLEANUP" = "True" ]; then
   oc delete -f $CNBI_MANIFEST > /dev/null
fi

# wait max 3 seconds until user presses
PROMPT_TIMEOUT=3
wait
