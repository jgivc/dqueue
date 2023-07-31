#!/bin/sh

INDEX=`cat operator_accounts.csv | grep -n $2 | awk -F: '{print $1-3}'`
test $INDEX || exit 10

sipp $1 -sf register.xml -inf operator_accounts.csv -infindex operator_accounts.csv 0 -m 1 -key line $INDEX
if [ $? -ne 0 ]; then
    exit 20
fi

sipp $1 -sf uas.xml -inf operator_accounts.csv -infindex operator_accounts.csv 0 -key line $INDEX -d $3 -trace_logs
# sipp $1 -sf uas.xml -inf operator_accounts.csv -infindex operator_accounts.csv 0 -key line $INDEX

