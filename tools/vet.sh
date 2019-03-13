#!/bin/bash

echo "Vetting code with \`go vet\`"
output=$(go vet "$@" 2>&1)
rval=$?
if [ -n "$output" ]; then
    echo "$(tput setaf 1)${output}$(tput sgr0)"
fi
exit $rval
